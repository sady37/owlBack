package httpapi

import (
	"net/http"
	"strings"
)

func (s *StubHandler) AdminAddresses(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/admin/api/v1/addresses":
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		case http.MethodPost:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"address_id":   "stub-address",
				"tenant_id":    "",
				"address_name": "stub",
				"is_active":    true,
			}))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/addresses/"):
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/addresses/")
		// allocate/*
		if strings.Contains(path, "/allocate/") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
			return
		}
		id := path
		if strings.Contains(id, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"address_id":   id,
				"tenant_id":    "",
				"address_name": "stub-" + id,
				"is_active":    true,
			}))
		case http.MethodPut:
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"address_id":   id,
				"tenant_id":    "",
				"address_name": "stub-" + id,
				"is_active":    true,
			}))
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, Ok[any](nil))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
