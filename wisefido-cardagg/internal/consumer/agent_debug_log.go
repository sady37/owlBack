package consumer

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

const agentDebugLogPath = "/home/wisefido/owl/.cursor/debug-d4fa48.log"

var agentDebugMu sync.Mutex

// #region agent log
func agentDebugLog(hypothesisID, location, message string, data map[string]any) {
	payload := map[string]any{
		"sessionId":    "d4fa48",
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	agentDebugMu.Lock()
	defer agentDebugMu.Unlock()
	f, err := os.OpenFile(agentDebugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(b, '\n'))
	_ = f.Close()
}

// #endregion
