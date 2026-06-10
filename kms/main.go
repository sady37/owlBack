package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"owl-common/crypto"
)

type KMS struct {
	mu             sync.RWMutex
	masterKeyPGP   []byte
	deploymentSalt []byte
}

type Archive struct {
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
	Salt      string `json:"salt"`
	Wrapped   string `json:"wrapped"`
}

func main() {
	initMode := flag.Bool("init", false, "Initialize KMS")
	recoverMode := flag.Bool("recover", false, "Recover KMS from archive")
	pin := flag.String("pin", "", "PIN for MW GPG encryption")
	archivePath := flag.String("archive", "", "Archive file path")
	mwToken := flag.String("mw-token", "", "MW cell that seals the existing archive (= its filename date's cell)")
	resealToken := flag.String("reseal-token", "", "MW cell to seal the NEW archive (today's cell)")
	masterPinFlag := flag.String("master-pin", "", "Master PIN — online /tenant-key auth only, never seals disk")
	socketPath := flag.String("socket", "/tmp/owl-kms.sock", "Unix socket path")
	vaultDir := flag.String("vault-dir", "owlBack/vault", "Vault directory")
	outDir := flag.String("out-dir", "owlBack/out", "Output directory")
	flag.Parse()

	os.MkdirAll(*vaultDir, 0750)
	os.MkdirAll(*outDir, 0750)
	os.MkdirAll("log", 0750)

	kms := &KMS{}

	if *initMode {
		if *pin == "" {
			log.Fatal("--pin required")
		}
		kms.doInit(*pin, *vaultDir, *outDir)
	} else if *recoverMode {
		mp := *masterPinFlag
		if mp == "" {
			mp = os.Getenv("MASTER_PIN")
		}
		if *archivePath == "" || *mwToken == "" || *resealToken == "" || mp == "" {
			log.Fatal("--archive, --mw-token (unseal: this archive's filename-date cell), --reseal-token (today's MW cell), --master-pin required")
		}
		kms.doRecover(mp, *mwToken, *resealToken, *archivePath, *vaultDir)
	} else {
		fmt.Println("Usage: owl-kms --init --pin <GPG-PIN>")
		fmt.Println("       owl-kms --recover --archive <path> --mw-token <unseal-cell> --reseal-token <today-cell> --master-pin <pin>")
		os.Exit(1)
	}
	kms.serve(*socketPath)
}

func (k *KMS) doInit(pin, vaultDir, outDir string) {
	masterKey := make([]byte, 32)
	mustRand(masterKey)
	masterPin := randAlphanumeric(8)
	k.deploymentSalt = make([]byte, 16)
	mustRand(k.deploymentSalt)

	wrapped, err := crypto.WrapMasterKey(masterKey, masterPin, "v1")
	if err != nil {
		log.Fatalf("wrap: %v", err)
	}
	k.masterKeyPGP = wrapped

	mwSeed := make([]byte, 24)
	mustRand(mwSeed)
	mwSeedB64 := base64.StdEncoding.EncodeToString(mwSeed)
	mwMD := renderMW(mwSeedB64)

	now := time.Now()
	mwToday := mwToken(mwSeedB64, int(now.Month()), isoWeekday(now.Weekday()))
	arcWrapped, _ := crypto.WrapMasterKey(masterKey, mwToday, "v1")
	arc := Archive{
		Version: "OWLKMS1", CreatedAt: now.UTC().Format(time.RFC3339),
		Salt: base64.StdEncoding.EncodeToString(k.deploymentSalt),
		Wrapped: base64.StdEncoding.EncodeToString(arcWrapped),
	}
	arcJSON, _ := json.MarshalIndent(arc, "", "  ")
	arcPath := filepath.Join(vaultDir, fmt.Sprintf("master_key-%s.json", now.Format("20060102")))
	os.WriteFile(arcPath, arcJSON, 0600)

	mwMDPath := filepath.Join(outDir, "mw.md")
	os.WriteFile(mwMDPath, []byte(mwMD), 0600)
	mwPGP := filepath.Join(outDir, "mw.pgp")
	cmd := exec.Command("gpg", "--batch", "--yes", "--passphrase-fd", "0", "--pinentry-mode", "loopback", "-c", "--cipher-algo", "AES256", "-o", mwPGP, mwMDPath)
	cmd.Stdin = strings.NewReader(pin)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("gpg: %v\n%s", err, out)
	}
	os.Remove(mwMDPath)
	zeroBytes(masterKey)
	auditLog("init", "", arcPath)

	fmt.Println("=== KMS Initialized ===")
	fmt.Printf("master_pin : %s\n", masterPin)
	fmt.Printf("MW         : %s\n", mwPGP)
	fmt.Printf("Archive    : %s\n", arcPath)
	fmt.Printf("MW_today (%s %s): %s\n", now.Month().String()[:3], now.Weekday().String()[:3], mwToday)
	fmt.Printf("\n  → Add to .env: MASTER_PIN=%s\n", masterPin)
	fmt.Printf("  → Add to .env: KMS_SOCKET=/tmp/owl-kms.sock\n")
}

func (k *KMS) doRecover(masterPin, unsealToken, resealToken, archivePath, vaultDir string) {
	data, err := os.ReadFile(archivePath)
	if err != nil {
		log.Fatalf("read archive %s: %v", archivePath, err)
	}
	var arc Archive
	if err := json.Unmarshal(data, &arc); err != nil {
		log.Fatalf("parse archive %s: %v", archivePath, err)
	}
	k.deploymentSalt, _ = base64.StdEncoding.DecodeString(arc.Salt)
	wrappedBytes, _ := base64.StdEncoding.DecodeString(arc.Wrapped)
	masterKey, _, err := crypto.UnwrapMasterKey(wrappedBytes, unsealToken)
	if err != nil {
		log.Fatalf("unseal failed (wrong MW cell for this archive's date?): %v", err)
	}
	defer zeroBytes(masterKey)

	// master_pin wraps the in-memory copy ONLY — it is the online /tenant-key auth gate,
	// never the disk seal. Disk stays MW-sealed so KMS death always requires the offline table.
	k.masterKeyPGP, _ = crypto.WrapMasterKey(masterKey, masterPin, "v1")

	// Reseal disk with today's MW cell: each recover rotates the seal so a once-seen cell
	// expires at the next restart. Filename date == the cell that seals it (see doc §4).
	reWrapped, _ := crypto.WrapMasterKey(masterKey, resealToken, "v1")
	newArc := Archive{Version: "OWLKMS1", CreatedAt: time.Now().UTC().Format(time.RFC3339), Salt: arc.Salt, Wrapped: base64.StdEncoding.EncodeToString(reWrapped)}
	newJSON, _ := json.MarshalIndent(newArc, "", "  ")
	newPath := filepath.Join(vaultDir, fmt.Sprintf("master_key-%s.json", time.Now().Format("20060102")))
	os.WriteFile(newPath, newJSON, 0600)
	cleanupArchives(vaultDir, 2)
	auditLog("recover", "", fmt.Sprintf("from=%s new=%s seal=mw", archivePath, newPath))
	fmt.Printf("KMS recovered. New archive (MW-sealed): %s\n", newPath)
}

func (k *KMS) serve(socketPath string) {
	os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	os.Chmod(socketPath, 0660)
	mux := http.NewServeMux()
	mux.HandleFunc("/tenant-key", k.handleTenantKey)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	fmt.Printf("KMS serving on %s\n", socketPath)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		ln.Close()
		os.Remove(socketPath)
		os.Exit(0)
	}()
	http.Serve(ln, mux)
}

func (k *KMS) handleTenantKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID  string `json:"tenant_id"`
		MasterPIN string `json:"master_pin"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.TenantID == "" || req.MasterPIN == "" {
		http.Error(w, "tenant_id and master_pin required", 400)
		return
	}
	k.mu.RLock()
	pgp, salt := k.masterKeyPGP, k.deploymentSalt
	k.mu.RUnlock()
	mk, _, err := crypto.UnwrapMasterKey(pgp, req.MasterPIN)
	if err != nil {
		http.Error(w, "invalid master_pin", 401)
		return
	}
	defer zeroBytes(mk)
	tk, _ := crypto.DeriveKey(mk, salt, []byte(req.TenantID))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"tenant_key": base64.StdEncoding.EncodeToString(tk)})
	auditLog("tenant-key", req.TenantID, "issued")
}

func renderMW(seed string) string {
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	wds := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	var b strings.Builder
	b.WriteString("# Matrix Wallet (MW)\n\n| Month ")
	for _, w := range wds { b.WriteString("| " + w + " ") }
	b.WriteString("|\n|---")
	for range wds { b.WriteString("|---") }
	b.WriteString("|\n")
	for mi := 1; mi <= 12; mi++ {
		b.WriteString("| " + months[mi-1] + " ")
		for wi := 1; wi <= 7; wi++ { b.WriteString("| " + mwToken(seed, mi, wi) + " ") }
		b.WriteString("|\n")
	}
	return b.String()
}

func mwToken(seed string, month, isoWd int) string {
	h := sha256.New()
	h.Write([]byte(seed))
	h.Write([]byte{byte(month), byte(isoWd)})
	tok := strings.TrimRight(base32.StdEncoding.EncodeToString(h.Sum(nil)), "=")
	if len(tok) > 6 { tok = tok[:6] }
	return strings.ToUpper(tok)
}

func isoWeekday(wd time.Weekday) int {
	if wd == time.Sunday { return 7 }
	return int(wd)
}

func mustRand(b []byte)    { rand.Read(b) }
func zeroBytes(b []byte)   { for i := range b { b[i] = 0 } }
func randAlphanumeric(n int) string {
	const ch = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"
	r := make([]byte, n)
	for i := range r { idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(ch)))); r[i] = ch[idx.Int64()] }
	return string(r)
}
func cleanupArchives(dir string, keep int) {
	m, _ := filepath.Glob(filepath.Join(dir, "master_key-*.json"))
	for i := 0; i < len(m)-keep; i++ { os.Remove(m[i]) }
}
func auditLog(action, tenant, detail string) {
	f, _ := os.OpenFile("log/unseal_audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if f != nil { defer f.Close(); host, _ := os.Hostname(); fmt.Fprintf(f, "%s\taction=%s\ttenant=%s\thost=%s\tdetail=%s\n", time.Now().UTC().Format(time.RFC3339), action, tenant, host, detail) }
}
