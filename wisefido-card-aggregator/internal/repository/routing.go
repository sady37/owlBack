package repository

import (
	"encoding/json"
	"fmt"
)

// ConvertUserListToUUIDArray 将 units.userList (JSONB) 转换为 UUID[] 数组
// userList 格式：["user_id1", "user_id2", ...] 或 [{"user_id": "..."}, ...]
func ConvertUserListToUUIDArray(userListJSON []byte) ([]string, error) {
	if len(userListJSON) == 0 || string(userListJSON) == "null" || string(userListJSON) == "[]" {
		return nil, nil
	}
	
	var userList interface{}
	if err := json.Unmarshal(userListJSON, &userList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal userList: %w", err)
	}
	
	var userIDs []string
	
	switch v := userList.(type) {
	case []interface{}:
		for _, item := range v {
			switch itemVal := item.(type) {
			case string:
				// 格式：["user_id1", "user_id2", ...]
				userIDs = append(userIDs, itemVal)
			case map[string]interface{}:
				// 格式：[{"user_id": "..."}, ...]
				if userID, ok := itemVal["user_id"].(string); ok {
					userIDs = append(userIDs, userID)
				}
			}
		}
	default:
		return nil, fmt.Errorf("unexpected userList format: %T", v)
	}
	
	return userIDs, nil
}

// ConvertGroupListToStringArray 将 units.groupList (JSONB) 转换为 VARCHAR[] 数组
// groupList 格式：["tag1", "tag2", ...] 或 [{"tag": "..."}, ...]
func ConvertGroupListToStringArray(groupListJSON []byte) ([]string, error) {
	if len(groupListJSON) == 0 || string(groupListJSON) == "null" || string(groupListJSON) == "[]" {
		return nil, nil
	}
	
	var groupList interface{}
	if err := json.Unmarshal(groupListJSON, &groupList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal groupList: %w", err)
	}
	
	var tags []string
	
	switch v := groupList.(type) {
	case []interface{}:
		for _, item := range v {
			switch itemVal := item.(type) {
			case string:
				// 格式：["tag1", "tag2", ...]
				tags = append(tags, itemVal)
			case map[string]interface{}:
				// 格式：[{"tag": "..."}, ...]
				if tag, ok := itemVal["tag"].(string); ok {
					tags = append(tags, tag)
				}
			}
		}
	default:
		return nil, fmt.Errorf("unexpected groupList format: %T", v)
	}
	
	return tags, nil
}

