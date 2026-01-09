package repository

import (
	"testing"
)

func TestConvertUserListToUUIDArray_SimpleArray(t *testing.T) {
	// 测试简单数组格式：["user_id1", "user_id2"]
	userListJSON := []byte(`["user-id-1", "user-id-2", "user-id-3"]`)

	result, err := ConvertUserListToUUIDArray(userListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 user IDs, got %d", len(result))
	}

	expected := []string{"user-id-1", "user-id-2", "user-id-3"}
	for i, id := range result {
		if id != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, id)
		}
	}
}

func TestConvertUserListToUUIDArray_ObjectArray(t *testing.T) {
	// 测试对象数组格式：[{"user_id": "..."}, ...]
	userListJSON := []byte(`[{"user_id": "user-id-1"}, {"user_id": "user-id-2"}]`)

	result, err := ConvertUserListToUUIDArray(userListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 user IDs, got %d", len(result))
	}

	expected := []string{"user-id-1", "user-id-2"}
	for i, id := range result {
		if id != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, id)
		}
	}
}

func TestConvertUserListToUUIDArray_EmptyArray(t *testing.T) {
	// 测试空数组
	userListJSON := []byte(`[]`)

	result, err := ConvertUserListToUUIDArray(userListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != nil && len(result) != 0 {
		t.Errorf("Expected empty or nil result, got %v", result)
	}
}

func TestConvertUserListToUUIDArray_Null(t *testing.T) {
	// 测试 null 值
	userListJSON := []byte(`null`)

	result, err := ConvertUserListToUUIDArray(userListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != nil && len(result) != 0 {
		t.Errorf("Expected empty or nil result, got %v", result)
	}
}

func TestConvertUserListToUUIDArray_InvalidJSON(t *testing.T) {
	// 测试无效 JSON
	userListJSON := []byte(`invalid json`)

	_, err := ConvertUserListToUUIDArray(userListJSON)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestConvertGroupListToStringArray_SimpleArray(t *testing.T) {
	// 测试简单数组格式：["tag1", "tag2"]
	groupListJSON := []byte(`["tag1", "tag2", "tag3"]`)

	result, err := ConvertGroupListToStringArray(groupListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(result))
	}

	expected := []string{"tag1", "tag2", "tag3"}
	for i, tag := range result {
		if tag != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, tag)
		}
	}
}

func TestConvertGroupListToStringArray_ObjectArray(t *testing.T) {
	// 测试对象数组格式：[{"tag": "..."}, ...]
	groupListJSON := []byte(`[{"tag": "tag1"}, {"tag": "tag2"}]`)

	result, err := ConvertGroupListToStringArray(groupListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(result))
	}

	expected := []string{"tag1", "tag2"}
	for i, tag := range result {
		if tag != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, tag)
		}
	}
}

func TestConvertGroupListToStringArray_EmptyArray(t *testing.T) {
	// 测试空数组
	groupListJSON := []byte(`[]`)

	result, err := ConvertGroupListToStringArray(groupListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != nil && len(result) != 0 {
		t.Errorf("Expected empty or nil result, got %v", result)
	}
}

func TestConvertGroupListToStringArray_Null(t *testing.T) {
	// 测试 null 值
	groupListJSON := []byte(`null`)

	result, err := ConvertGroupListToStringArray(groupListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != nil && len(result) != 0 {
		t.Errorf("Expected empty or nil result, got %v", result)
	}
}

func TestConvertGroupListToStringArray_InvalidJSON(t *testing.T) {
	// 测试无效 JSON
	groupListJSON := []byte(`invalid json`)

	_, err := ConvertGroupListToStringArray(groupListJSON)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestConvertUserListToUUIDArray_MixedFormat(t *testing.T) {
	// 测试混合格式（不应该出现，但测试容错性）
	userListJSON := []byte(`["user-id-1", {"user_id": "user-id-2"}]`)

	result, err := ConvertUserListToUUIDArray(userListJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 应该能处理混合格式
	if len(result) < 1 {
		t.Fatalf("Expected at least 1 user ID, got %d", len(result))
	}
}

// 基准测试
func BenchmarkConvertUserListToUUIDArray(b *testing.B) {
	userListJSON := []byte(`["user-id-1", "user-id-2", "user-id-3", "user-id-4", "user-id-5"]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertUserListToUUIDArray(userListJSON)
	}
}

func BenchmarkConvertGroupListToStringArray(b *testing.B) {
	groupListJSON := []byte(`["tag1", "tag2", "tag3", "tag4", "tag5"]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertGroupListToStringArray(groupListJSON)
	}
}
