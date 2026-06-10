// Ad-hoc PHI decrypt verify — not for prod path.
//
// 适配当前 schema：resident_phi.resident_id (INET PK)，*_enc BYTEA 列存
// owl-common envelope (nonce||ct||tag)；_iv/_tag 列遗留未用；无 tenant_id/明文列。
// tenant_id (AAD + KMS key 派生) 取自 residents.tenant_id::text（/48 CIDR 形）。
//
// Usage:
//   MASTER_PIN=... go run ./cmd/verify_phi                    # round-trip：任一行 first_name_enc 解密
//   MASTER_PIN=... go run ./cmd/verify_phi <nickname-or-inet> # 指定 resident，dump 所有非空 *_enc
//
// 读 MASTER_PIN from env（不打印）。向 KMS 要 tenant_key，解 PG resident_phi.*_enc。
// GCM 认证标签保证：key/AAD 不对 → Open 失败，所以"解出可读明文"即端到端通路活。
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

func main() {
	pin := os.Getenv("MASTER_PIN")
	if pin == "" {
		fmt.Fprintln(os.Stderr, "MASTER_PIN env required")
		os.Exit(2)
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@127.0.0.1:5432/owl_v2?sslmode=disable"
	}
	sock := os.Getenv("KMS_SOCKET")
	if sock == "" {
		sock = "/tmp/owl-kms.sock"
	}

	db, err := sql.Open("postgres", dbURL)
	must(err, "db open")
	defer db.Close()

	fmt.Printf("[%s] verify_phi\n", time.Now().Format("2006-01-02 15:04:05"))

	if len(os.Args) >= 2 {
		runNamed(db, sock, pin, os.Args[1])
		return
	}
	runRoundTrip(db, sock, pin)
}

// runRoundTrip: 任选一行非空 first_name_enc，用 KMS 派生 key 解密。
// 无明文列可比对，GCM Open 成功（无 error）即证明 key + AAD + 通路全对。
func runRoundTrip(db *sql.DB, sock, pin string) {
	var rid, tid string
	var ct []byte
	err := db.QueryRow(`
		SELECT resident_id::text,
		       network(set_masklen(resident_id, 48))::text,
		       first_name_enc
		  FROM resident_phi
		 WHERE first_name_enc IS NOT NULL AND length(first_name_enc) > 12
		 LIMIT 1`).Scan(&rid, &tid, &ct)
	must(err, "no eligible round-trip row")

	key := tenantKey(sock, pin, tid)
	pt, err := kmscrypto.DecryptWithDataKey(key, []byte(tid), ct)
	must(err, "decrypt (key/AAD/通路异常)")

	fmt.Printf("tenant=%s resident=%s field=first_name plain=%q cipher=%dB\n", tid, rid, string(pt), len(ct))
	fmt.Println("OK — END-TO-END decrypt 成功（KMS 在线 + tenant_key 派生正确）")
}

// runNamed: 按 nickname 或 resident_id(INET) 定位，dump 所有非空 *_enc 字段解密结果。
func runNamed(db *sql.DB, sock, pin, target string) {
	var rid, tid, nickname string
	err := db.QueryRow(`
		SELECT resident_id::text,
		       network(set_masklen(resident_id, 48))::text,
		       COALESCE(nickname,'')
		  FROM residents
		 WHERE resident_id::text = $1 OR nickname = $1
		 LIMIT 1`, target).Scan(&rid, &tid, &nickname)
	must(err, "resident not found: "+target)
	fmt.Printf("tenant=%s resident=%s nickname=%s\n\n", tid, rid, nickname)

	encCols := encColumns(db)
	if len(encCols) == 0 {
		fmt.Fprintln(os.Stderr, "no *_enc columns in resident_phi")
		os.Exit(1)
	}

	cols := strings.Join(encCols, ",")
	row := db.QueryRow(`SELECT `+cols+` FROM resident_phi WHERE resident_id = $1::INET`, rid)
	blobs := make([][]byte, len(encCols))
	dest := make([]any, len(encCols))
	for i := range encCols {
		dest[i] = &blobs[i]
	}
	must(row.Scan(dest...), "scan PHI")

	key := tenantKey(sock, pin, tid)
	okCount, failCount, emptyCount := 0, 0, 0
	for i, col := range encCols {
		field := strings.TrimSuffix(col, "_enc")
		ct := blobs[i]
		if len(ct) == 0 {
			emptyCount++
			continue
		}
		pt, err := kmscrypto.DecryptWithDataKey(key, []byte(tid), ct)
		if err != nil {
			fmt.Printf("  %-28s  FAIL: %v\n", field, err)
			failCount++
			continue
		}
		fmt.Printf("  %-28s = %-30q  (cipher %dB)\n", field, string(pt), len(ct))
		okCount++
	}
	fmt.Printf("\nfields: %d decrypted OK, %d failed, %d empty/null\n", okCount, failCount, emptyCount)
	if failCount == 0 && okCount > 0 {
		fmt.Println("OK — 所有非空 PHI 字段解密成功")
	} else if okCount == 0 {
		fmt.Println("WARN — 该 resident 无非空 *_enc 字段")
	} else {
		os.Exit(1)
	}
}

func encColumns(db *sql.DB) []string {
	rows, err := db.Query(`
		SELECT column_name FROM information_schema.columns
		 WHERE table_name = 'resident_phi' AND column_name LIKE '%\_enc'
		 ORDER BY ordinal_position`)
	must(err, "list _enc columns")
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var c string
		must(rows.Scan(&c), "scan col")
		cols = append(cols, c)
	}
	return cols
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
