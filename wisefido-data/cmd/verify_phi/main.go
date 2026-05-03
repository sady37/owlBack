// Ad-hoc PHI decrypt verify — not for prod path.
//
// Usage:
//   MASTER_PIN=... go run ./cmd/verify_phi                    # default: 1st row with non-empty first_name+_enc
//   MASTER_PIN=... go run ./cmd/verify_phi <nickname-or-uuid> # named resident, dump all non-empty *_enc + note
//
// Reads MASTER_PIN from env (never prints it). Asks KMS for tenant_key, decrypts
// real ciphertext from PG resident_phi.*_enc, prints plaintext + cipher byte length.
package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	kmscrypto "owl-common/crypto"

	_ "github.com/lib/pq"
)

var phiFields = []string{
	"first_name", "last_name", "gender", "date_of_birth",
	"resident_phone", "resident_email",
	"weight_lb", "height_ft", "height_in", "mobility_level",
	"tremor_status", "mobility_aid", "adl_assistance", "comm_status",
	"has_hypertension", "has_hyperlipaemia", "has_hyperglycaemia",
	"has_stroke_history", "has_paralysis", "has_alzheimer",
	"medical_history",
	"home_address_street", "home_address_city", "home_address_state",
	"home_address_postal_code", "plus_code",
}

func main() {
	pin := os.Getenv("MASTER_PIN")
	if pin == "" {
		fmt.Fprintln(os.Stderr, "MASTER_PIN env required")
		os.Exit(2)
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@127.0.0.1:5432/owlrd?sslmode=disable"
	}
	sock := os.Getenv("KMS_SOCKET")
	if sock == "" {
		sock = "/tmp/owl-kms.sock"
	}

	db, err := sql.Open("postgres", dbURL)
	must(err, "db open")
	defer db.Close()

	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] verify_phi\n", now)

	if len(os.Args) >= 2 {
		runNamed(db, sock, pin, os.Args[1])
		return
	}
	runRoundTrip(db, sock, pin)
}

// runRoundTrip: pick any row with non-empty first_name+first_name_enc, verify decrypt matches plaintext.
func runRoundTrip(db *sql.DB, sock, pin string) {
	var tid, pid, plain string
	var ct []byte
	err := db.QueryRow(`SELECT tenant_id::text, phi_id::text, first_name, first_name_enc
		FROM resident_phi
		WHERE first_name IS NOT NULL AND first_name <> ''
		  AND first_name_enc IS NOT NULL AND length(first_name_enc) > 30
		LIMIT 1`).Scan(&tid, &pid, &plain, &ct)
	must(err, "no eligible round-trip row")

	key := tenantKey(sock, pin, tid)
	pt, err := kmscrypto.DecryptWithDataKey(key, []byte(tid), ct)
	must(err, "decrypt")
	if string(pt) == plain {
		fmt.Printf("tenant=%s phi=%s field=first_name plain=%q cipher=%dB\n", tid, pid[:8], plain, len(ct))
		fmt.Println("OK — END-TO-END decrypt matches plaintext")
		return
	}
	fmt.Printf("FAIL — got %q, expected %q\n", string(pt), plain)
	os.Exit(1)
}

// runNamed: resolve nickname or UUID, dump every non-empty *_enc + the residents.note plaintext.
func runNamed(db *sql.DB, sock, pin, target string) {
	var tid, rid, nickname, note string
	var noteN sql.NullString
	q := `SELECT t.tenant_id::text, t.tenant_name, r.resident_id::text, r.nickname, r.note
		FROM residents r JOIN tenants t USING(tenant_id)
		WHERE r.resident_id::text=$1 OR r.nickname=$1
		LIMIT 1`
	var tname string
	err := db.QueryRow(q, target).Scan(&tid, &tname, &rid, &nickname, &noteN)
	must(err, "resident not found: "+target)
	if noteN.Valid {
		note = noteN.String
	}
	fmt.Printf("tenant=%s/%s nickname=%s resident=%s\n", tname, tid[:8], nickname, rid[:8])
	fmt.Printf("residents.note (plaintext, not PHI) = %q\n\n", note)

	key := tenantKey(sock, pin, tid)

	cols := make([]string, 0, len(phiFields)*2)
	for _, f := range phiFields {
		cols = append(cols, "COALESCE("+f+"::text,'')", f+"_enc")
	}
	row := db.QueryRow(`SELECT `+strings.Join(cols, ",")+` FROM resident_phi WHERE resident_id=$1`, rid)
	dest := make([]any, len(cols))
	plains := make([]string, len(phiFields))
	cts := make([][]byte, len(phiFields))
	for i := range phiFields {
		dest[2*i] = &plains[i]
		dest[2*i+1] = &cts[i]
	}
	must(row.Scan(dest...), "scan PHI")

	okCount, failCount, emptyCount := 0, 0, 0
	for i, f := range phiFields {
		ct := cts[i]
		if len(ct) == 0 {
			emptyCount++
			continue
		}
		pt, err := kmscrypto.DecryptWithDataKey(key, []byte(tid), ct)
		if err != nil {
			fmt.Printf("  %-26s  FAIL: %v\n", f, err)
			failCount++
			continue
		}
		fmt.Printf("  %-26s = %-30q  (cipher %dB, plaintext_col=%q)\n", f, string(pt), len(ct), plains[i])
		okCount++
	}
	fmt.Printf("\nfields: %d decrypted OK, %d failed, %d empty/null\n", okCount, failCount, emptyCount)
	if failCount == 0 && okCount > 0 {
		fmt.Println("OK — all non-empty PHI fields decrypted successfully")
	} else if okCount == 0 {
		fmt.Println("WARN — no non-empty *_enc fields to verify")
	} else {
		os.Exit(1)
	}
}

func tenantKey(sock, pin, tid string) []byte {
	body, _ := json.Marshal(map[string]string{"tenant_id": tid, "master_pin": pin})
	cli := &http.Client{
		Transport: &http.Transport{DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", sock)
		}},
		Timeout: 5 * time.Second,
	}
	resp, err := cli.Post("http://kms/tenant-key", "application/json", strings.NewReader(string(body)))
	must(err, "KMS dial")
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "KMS HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}
	var r struct {
		TenantKey string `json:"tenant_key"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	key, _ := base64.StdEncoding.DecodeString(r.TenantKey)
	return key
}

func must(err error, ctx string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", ctx, err)
		os.Exit(1)
	}
}
