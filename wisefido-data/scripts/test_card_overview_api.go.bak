// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// 测试配置
var (
	baseURL  = getEnv("BASE_URL", "http://localhost:8080")
	tenantID = getEnv("TENANT_ID", "00000000-0000-0000-0000-000000000001")
	userID   = getEnv("USER_ID", "00000000-0000-0000-0000-000000000002")
	userType = getEnv("USER_TYPE", "resident")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	fmt.Println("==========================================")
	fmt.Println("Card Overview API 测试")
	fmt.Println("==========================================")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Tenant ID: %s\n", tenantID)
	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("User Type: %s\n", userType)
	fmt.Println()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 测试 1: 基本查询
	testBasicQuery(client)

	// 测试 2: 搜索查询
	testSearchQuery(client)

	// 测试 3: 按卡片类型过滤
	testCardTypeFilter(client)

	// 测试 4: 组合查询
	testCombinedQuery(client)

	// 测试 5: 排序
	testSorting(client)

	fmt.Println("==========================================")
	fmt.Println("测试完成")
	fmt.Println("==========================================")
}

func testBasicQuery(client *http.Client) {
	fmt.Println("测试 1: 基本查询")
	fmt.Println("----------------------------------------")
	url := fmt.Sprintf("%s/admin/api/v1/card-overview?tenant_id=%s", baseURL, tenantID)
	makeRequest(client, "GET", url, nil)
	fmt.Println()
}

func testSearchQuery(client *http.Client) {
	fmt.Println("测试 2: 搜索查询 (search=Test)")
	fmt.Println("----------------------------------------")
	url := fmt.Sprintf("%s/admin/api/v1/card-overview?tenant_id=%s&search=Test", baseURL, tenantID)
	makeRequest(client, "GET", url, nil)
	fmt.Println()
}

func testCardTypeFilter(client *http.Client) {
	fmt.Println("测试 3: 按卡片类型过滤 (card_type=ActiveBed)")
	fmt.Println("----------------------------------------")
	url := fmt.Sprintf("%s/admin/api/v1/card-overview?tenant_id=%s&card_type=ActiveBed", baseURL, tenantID)
	makeRequest(client, "GET", url, nil)
	fmt.Println()
}

func testCombinedQuery(client *http.Client) {
	fmt.Println("测试 4: 组合查询 (search=Test&card_type=ActiveBed)")
	fmt.Println("----------------------------------------")
	url := fmt.Sprintf("%s/admin/api/v1/card-overview?tenant_id=%s&search=Test&card_type=ActiveBed", baseURL, tenantID)
	makeRequest(client, "GET", url, nil)
	fmt.Println()
}

func testSorting(client *http.Client) {
	fmt.Println("测试 5: 排序 (sort=card_name&direction=desc)")
	fmt.Println("----------------------------------------")
	url := fmt.Sprintf("%s/admin/api/v1/card-overview?tenant_id=%s&sort=card_name&direction=desc", baseURL, tenantID)
	makeRequest(client, "GET", url, nil)
	fmt.Println()
}

func makeRequest(client *http.Client, method, url string, body io.Reader) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("X-User-Id", userID)
	req.Header.Set("X-User-Type", userType)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("状态码: %d\n", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	// 尝试格式化 JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, respBody, "", "  "); err != nil {
		fmt.Printf("响应内容（非 JSON）:\n%s\n", string(respBody))
		return
	}

	fmt.Printf("响应内容:\n%s\n", prettyJSON.String())

	// 解析响应检查结构
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err == nil {
		if code, ok := result["code"].(float64); ok {
			if code == 2000 {
				fmt.Println("✅ 请求成功")
				if data, ok := result["result"].(map[string]any); ok {
					if items, ok := data["items"].([]any); ok {
						fmt.Printf("✅ 返回 %d 个卡片\n", len(items))
					}
					if pagination, ok := data["pagination"].(map[string]any); ok {
						if total, ok := pagination["total"].(float64); ok {
							fmt.Printf("✅ 总数: %.0f\n", total)
						}
					}
				}
			} else {
				fmt.Printf("⚠️  请求返回错误码: %.0f\n", code)
				if msg, ok := result["message"].(string); ok {
					fmt.Printf("错误信息: %s\n", msg)
				}
			}
		}
	}
}

