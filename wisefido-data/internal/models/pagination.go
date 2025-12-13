package models

// BackendPagination 与 owlFront/src/api/model/pagination.ts 保持一致
type BackendPagination struct {
	Size      int    `json:"size"`
	Page      int    `json:"page"`
	Count     int    `json:"count"`
	Sort      string `json:"sort"`
	Direction int    `json:"direction"`
}


