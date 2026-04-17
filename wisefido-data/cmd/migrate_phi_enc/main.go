package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	kmscrypto "owl-common/crypto"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" { dbURL = "postgres://postgres:postgres@127.0.0.1:5432/owlrd?sslmode=disable" }
	kmsSocket := os.Getenv("KMS_SOCKET")
	if kmsSocket == "" { kmsSocket = "/tmp/owl-kms.sock" }
	masterPin := os.Getenv("MASTER_PIN")
	if masterPin == "" { log.Fatal("MASTER_PIN env required") }

	db, err := sql.Open("postgres", dbURL)
	if err != nil { log.Fatalf("open db: %v", err) }
	defer db.Close()

	rows, _ := db.Query("SELECT DISTINCT tenant_id::text FROM resident_phi")
	var tids []string
	for rows.Next() { var t string; rows.Scan(&t); tids = append(tids, t) }
	rows.Close()
	log.Printf("Found %d tenants", len(tids))

	keys := map[string][]byte{}
	for _, t := range tids {
		k, err := getTenantKey(kmsSocket, masterPin, t)
		if err != nil { log.Fatalf("get key %s: %v", t, err) }
		keys[t] = k
	}

	dataRows, _ := db.Query(`SELECT phi_id::text,tenant_id::text,
		COALESCE(first_name,''),COALESCE(last_name,''),COALESCE(gender,''),date_of_birth,
		COALESCE(resident_phone,''),COALESCE(resident_email,''),
		weight_lb,height_ft,height_in,mobility_level,
		COALESCE(tremor_status,''),COALESCE(mobility_aid,''),COALESCE(adl_assistance,''),COALESCE(comm_status,''),
		COALESCE(has_hypertension,false),COALESCE(has_hyperlipaemia,false),COALESCE(has_hyperglycaemia,false),
		COALESCE(has_stroke_history,false),COALESCE(has_paralysis,false),COALESCE(has_alzheimer,false),
		COALESCE(medical_history,''),
		COALESCE(home_address_street,''),COALESCE(home_address_city,''),COALESCE(home_address_state,''),
		COALESCE(home_address_postal_code,''),COALESCE(plus_code,'')
	FROM resident_phi WHERE first_name_enc IS NULL ORDER BY tenant_id`)
	defer dataRows.Close()

	n := 0
	for dataRows.Next() {
		var pid, tid string
		var fn, ln, g string; var dob sql.NullTime
		var ph, em string
		var wt, hf, hi sql.NullFloat64; var ml sql.NullInt64
		var ts, ma, ad, cs string
		var hy, hl, hg, st, pa, al bool
		var mh, as_, ac, astate, ap, pc string

		dataRows.Scan(&pid, &tid, &fn, &ln, &g, &dob, &ph, &em, &wt, &hf, &hi, &ml,
			&ts, &ma, &ad, &cs, &hy, &hl, &hg, &st, &pa, &al, &mh, &as_, &ac, &astate, &ap, &pc)

		key, aad := keys[tid], []byte(tid)
		enc := make([][]byte, 26)
		enc[0] = encS(key, aad, fn); enc[1] = encS(key, aad, ln); enc[2] = encS(key, aad, g)
		if dob.Valid { enc[3] = encS(key, aad, dob.Time.Format("2006-01-02")) }
		enc[4] = encS(key, aad, ph); enc[5] = encS(key, aad, em)
		if wt.Valid { enc[6] = encS(key, aad, strconv.FormatFloat(wt.Float64, 'f', 2, 64)) }
		if hf.Valid { enc[7] = encS(key, aad, strconv.FormatFloat(hf.Float64, 'f', 2, 64)) }
		if hi.Valid { enc[8] = encS(key, aad, strconv.FormatFloat(hi.Float64, 'f', 2, 64)) }
		if ml.Valid { enc[9] = encS(key, aad, strconv.FormatInt(ml.Int64, 10)) }
		enc[10] = encS(key, aad, ts); enc[11] = encS(key, aad, ma); enc[12] = encS(key, aad, ad); enc[13] = encS(key, aad, cs)
		enc[14] = encS(key, aad, strconv.FormatBool(hy)); enc[15] = encS(key, aad, strconv.FormatBool(hl))
		enc[16] = encS(key, aad, strconv.FormatBool(hg)); enc[17] = encS(key, aad, strconv.FormatBool(st))
		enc[18] = encS(key, aad, strconv.FormatBool(pa)); enc[19] = encS(key, aad, strconv.FormatBool(al))
		enc[20] = encS(key, aad, mh); enc[21] = encS(key, aad, as_); enc[22] = encS(key, aad, ac)
		enc[23] = encS(key, aad, astate); enc[24] = encS(key, aad, ap); enc[25] = encS(key, aad, pc)

		db.ExecContext(context.Background(), `UPDATE resident_phi SET
			first_name_enc=$1,last_name_enc=$2,gender_enc=$3,date_of_birth_enc=$4,
			resident_phone_enc=$5,resident_email_enc=$6,weight_lb_enc=$7,height_ft_enc=$8,
			height_in_enc=$9,mobility_level_enc=$10,tremor_status_enc=$11,mobility_aid_enc=$12,
			adl_assistance_enc=$13,comm_status_enc=$14,has_hypertension_enc=$15,has_hyperlipaemia_enc=$16,
			has_hyperglycaemia_enc=$17,has_stroke_history_enc=$18,has_paralysis_enc=$19,has_alzheimer_enc=$20,
			medical_history_enc=$21,home_address_street_enc=$22,home_address_city_enc=$23,
			home_address_state_enc=$24,home_address_postal_code_enc=$25,plus_code_enc=$26
		WHERE phi_id=$27`,
			enc[0], enc[1], enc[2], enc[3], enc[4], enc[5], enc[6], enc[7], enc[8], enc[9],
			enc[10], enc[11], enc[12], enc[13], enc[14], enc[15], enc[16], enc[17], enc[18], enc[19],
			enc[20], enc[21], enc[22], enc[23], enc[24], enc[25], pid)
		n++
		log.Printf("Migrated %s (%s)", pid[:8], fn)
	}
	log.Printf("Done: %d rows", n)
}

func encS(key, aad []byte, s string) []byte {
	if s == "" { return nil }
	ct, _ := kmscrypto.EncryptWithDataKey(key, aad, []byte(s))
	return ct
}

func getTenantKey(sock, pin, tid string) ([]byte, error) {
	body, _ := json.Marshal(map[string]string{"tenant_id": tid, "master_pin": pin})
	c := &http.Client{Transport: &http.Transport{DialContext: func(_ context.Context, _, _ string) (net.Conn, error) { return net.Dial("unix", sock) }}, Timeout: 5 * time.Second}
	resp, err := c.Post("http://kms/tenant-key", "application/json", strings.NewReader(string(body)))
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != 200 { return nil, fmt.Errorf("K returned %d", resp.StatusCode) }
	var r struct{ TenantKey string `json:"tenant_key"` }
	json.NewDecoder(resp.Body).Decode(&r)
	return base64.StdEncoding.DecodeString(r.TenantKey)
}
