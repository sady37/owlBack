// apns-test：从 apns_devices 取一条 staff 活跃 token 发测试推送，或 -token/-env 手动指定。
// 用法（与 wisefido-data 相同工作目录与 CONFIG_PATH、APNS_* 环境变量）：
//
//	cd owlBack/wisefido-data && CONFIG_PATH=config.yaml \
//	  APNS_KEY_ID=... APNS_TEAM_ID=... APNS_BUNDLE_ID=... APNS_P8_KEY=... \
//	  go run ./cmd/apns-test
//
//	go run ./cmd/apns-test -tenant <uuid>
//	go run ./cmd/apns-test -tenant <uuid> -env sandbox   // 覆盖 apns_devices.environment（未上架调试常用）
//	go run ./cmd/apns-test -token <hex> -env sandbox
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"wisefido-data/internal/config"
	"wisefido-data/internal/notify"

	"owl-common/database"
)

func main() {
	tenant := flag.String("tenant", "", "optional tenant_id when picking from DB")
	tokenArg := flag.String("token", "", "device token hex (skips DB pick)")
	envArg := flag.String("env", "", "sandbox|production (required with -token; optional with -tenant to override DB column)")
	tryBoth := flag.Bool("try-both", true, "on BadDeviceToken retry opposite env")
	flag.Parse()

	cfg := config.Load()
	if !cfg.DBEnabled {
		log.Fatal("db_enabled is false in config")
	}
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	kid := strings.TrimSpace(os.Getenv("APNS_KEY_ID"))
	if kid == "" {
		log.Fatal("APNS_KEY_ID empty")
	}
	sender, err := notify.NewAPNSSender(
		kid,
		strings.TrimSpace(os.Getenv("APNS_TEAM_ID")),
		strings.TrimSpace(os.Getenv("APNS_BUNDLE_ID")),
		os.Getenv("APNS_P8_KEY"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctxPick, cancelPick := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPick()

	var token, env string
	if strings.TrimSpace(*tokenArg) != "" {
		token = normalizeToken(*tokenArg)
		env = strings.TrimSpace(*envArg)
		if env != "sandbox" && env != "production" {
			log.Fatal("-env must be sandbox or production when using -token")
		}
	} else {
		var row *sql.Row
		if strings.TrimSpace(*tenant) != "" {
			row = db.QueryRowContext(ctxPick, `
				SELECT device_token, environment
				FROM apns_devices
				WHERE is_active = TRUE AND user_type = 'staff' AND device_token <> ''
				  AND tenant_id = $1::uuid
				ORDER BY updated_at DESC
				LIMIT 1
			`, strings.TrimSpace(*tenant))
		} else {
			row = db.QueryRowContext(ctxPick, `
				SELECT device_token, environment
				FROM apns_devices
				WHERE is_active = TRUE AND user_type = 'staff' AND device_token <> ''
				ORDER BY updated_at DESC
				LIMIT 1
			`)
		}
		if err := row.Scan(&token, &env); err != nil {
			log.Fatalf("pick row: %v (need staff apns_devices or use -token)", err)
		}
		token = normalizeToken(token)
		env = strings.TrimSpace(env)
		if env == "" {
			env = "production"
		}
		_, _ = fmt.Fprintf(os.Stderr, "picked environment=%q token_prefix=%s\n", env, tokenPrefix(token))
	}

	if strings.TrimSpace(*tokenArg) == "" && strings.TrimSpace(*envArg) != "" {
		e := strings.TrimSpace(*envArg)
		if e != "sandbox" && e != "production" {
			log.Fatal("-env must be sandbox or production")
		}
		env = e
		_, _ = fmt.Fprintf(os.Stderr, "env overridden by -env flag: %q\n", env)
	}

	badge := 1
	payload := notify.APNSPayload{
		APS: notify.APSDict{
			Alert:            notify.APSAlert{Title: "APNs test", Body: "wisefido-data apns-test"},
			Sound:            "default",
			Badge:            &badge,
			ContentAvailable: 1,
		},
	}

	send := func(e string) error {
		c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return sender.Send(c, token, e, payload)
	}

	err = send(env)
	if err != nil && *tryBoth && strings.Contains(err.Error(), "BadDeviceToken") {
		alt := "production"
		if env == "production" {
			alt = "sandbox"
		}
		_, _ = fmt.Fprintf(os.Stderr, "first send failed (%v), retry env=%q\n", err, alt)
		err = send(alt)
		if err == nil {
			env = alt
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ok sent env=%s token_prefix=%s\n", env, tokenPrefix(token))
}

func normalizeToken(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

func tokenPrefix(t string) string {
	if len(t) >= 12 {
		return t[:12]
	}
	return t
}
