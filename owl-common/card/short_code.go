package card

import "crypto/sha256"

// sha256SumImpl — 单独文件，让 card_types.go 不直接 import crypto，避免 type 文件臃肿
func sha256SumImpl(b []byte) [32]byte { return sha256.Sum256(b) }
