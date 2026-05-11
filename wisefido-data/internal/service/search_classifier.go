package service

import (
	"regexp"
	"strings"
)

type SearchType int

const (
	SearchTypeUnknown SearchType = iota
	SearchTypeNickname
	SearchTypeAccount
	SearchTypeEmail
	SearchTypePhone
)

// ClassifySearch 根据搜索词自动识别搜索类型
// - Email: 包含 @ 符号的有效 email 格式
// - Phone: 纯数字或含 +- 的电话号码格式
// - Account: 短字符串（通常是 user/resident account）
// - Nickname: 其他（通常是昵称）
func ClassifySearch(searchTerm string) SearchType {
	searchTerm = strings.TrimSpace(searchTerm)
	if searchTerm == "" {
		return SearchTypeUnknown
	}

	// 1. Email 匹配：包含 @ 且符合基本 email 格式
	if strings.Contains(searchTerm, "@") {
		emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
		if emailRegex.MatchString(searchTerm) {
			return SearchTypeEmail
		}
	}

	// 2. Phone 匹配：纯数字、+开头、含-或空格的电话格式
	phoneRegex := regexp.MustCompile(`^[\d\s+\-()]+$`)
	if phoneRegex.MatchString(searchTerm) && len(strings.TrimSpace(strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, searchTerm))) > 0 {
		return SearchTypePhone
	}

	// 3. Account 匹配：短字符串（<20 字符）、不含空格、通常是 alphanumeric + 下划线/短线
	accountRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,20}$`)
	if accountRegex.MatchString(searchTerm) {
		return SearchTypeAccount
	}

	// 4. 默认：Nickname（含空格或特殊字符的人名）
	return SearchTypeNickname
}

// GetSearchTypeDescription 返回搜索类型的描述
func GetSearchTypeDescription(searchType SearchType) string {
	switch searchType {
	case SearchTypeEmail:
		return "email"
	case SearchTypePhone:
		return "phone"
	case SearchTypeAccount:
		return "account"
	case SearchTypeNickname:
		return "nickname"
	default:
		return "unknown"
	}
}
