package service

import (
	"crypto/sha256"
	"encoding/hex"
)

// GeneratePasswordHash 生成密码哈希（用于测试用户和密码初始化）
// 格式：SHA256(password) - 无盐，与数据库现有格式相同
func GeneratePasswordHash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
