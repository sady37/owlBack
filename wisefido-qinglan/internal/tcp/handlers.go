package tcp

// OTA-focused message handler (simplified from ota-main)

import (
	"fmt"
	"log"
	"net"

	pb "wisefido-qinglan/proto/gen"

	"google.golang.org/protobuf/proto"
)

// OTAProgressCallback OTA progress callback
type OTAProgressCallback func(uid string, progress int, message string)

// OnRegisterCallback called when a device registers via TCP, to update device_store
type OnRegisterCallback func(uid, deviceType, sfVer, hwVer string)

// MsgType constants
const (
	TypeGetServer     byte = 1
	TypeGetServerResp byte = 2
	TypeRegister      byte = 3
	TypeRegisterResp  byte = 4
	TypeHeartbeat     byte = 7
	TypeHeartbeatResp byte = 8
	TypeOTAPush       byte = 16
	TypeOTAResp       byte = 17
	TypeOTAProgress   byte = 18
)

// HandleFrame handles a single TCP frame
func HandleFrame(conn net.Conn, frame *Frame, sm *SessionManager, serverAddr string, serverPort uint32, onProgress OTAProgressCallback, onRegister OnRegisterCallback) {
	switch frame.Type {

	case TypeGetServer:
		req := &pb.GetServerReq{}
		if err := proto.Unmarshal(frame.Data, req); err != nil {
			log.Printf("[TCP] type 1 decode failed: %v", err)
			return
		}
		log.Printf("[TCP] dispatch request: UID=%s type=%s", req.Uid, req.Type)
		resp := &pb.GetServerResponse{Seq: 2, Result: 0, Server: serverAddr, Port: serverPort}
		sendProto(conn, TypeGetServerResp, resp)

	case TypeRegister:
		req := &pb.RegisterReq{}
		if err := proto.Unmarshal(frame.Data, req); err != nil {
			log.Printf("[TCP] type 3 decode failed: %v", err)
			return
		}
		log.Printf("[TCP] register: UID=%s type=%s ver=%s HW=%s", req.Uid, req.Type, req.Sfver, req.Hwver)
		sm.Connect(conn, req.Uid, req.Type, req.Sfver, req.Hwver)
		resp := &pb.RegisterResponse{Seq: 4, Result: 0}
		sendProto(conn, TypeRegisterResp, resp)
		// Callback: update device_store with firmware version + online status
		if onRegister != nil {
			onRegister(req.Uid, req.Type, req.Sfver, req.Hwver)
		}

	case TypeHeartbeat:
		sm.UpdateHeartbeat(conn)
		resp := &pb.CommonMessage{Seq: 8}
		sendProto(conn, TypeHeartbeatResp, resp)

	case TypeOTAResp:
		msg := &pb.OtaResponse{}
		if err := proto.Unmarshal(frame.Data, msg); err != nil {
			log.Printf("[TCP] type 17 decode failed: %v", err)
			return
		}
		uid := sm.GetUIDByConn(conn)
		if msg.Result == 0 {
			log.Printf("[TCP] OTA response: UID=%s device accepted OTA", uid)
			if onProgress != nil {
				onProgress(uid, 0, "device accepted OTA")
			}
		} else {
			log.Printf("[TCP] OTA response: UID=%s rejected result=%d errmsg=%s", uid, msg.Result, msg.Errmsg)
			if onProgress != nil {
				onProgress(uid, -1, fmt.Sprintf("device rejected: %s", msg.Errmsg))
			}
		}

	case TypeOTAProgress:
		msg := &pb.OTAProgress{}
		if err := proto.Unmarshal(frame.Data, msg); err != nil {
			log.Printf("[TCP] type 18 decode failed: %v", err)
			return
		}
		uid := sm.GetUIDByConn(conn)
		progressMsg := ""
		switch msg.Progress {
		case -1:
			log.Printf("[TCP] OTA failed: UID=%s %s", uid, msg.ErrMsg)
			progressMsg = fmt.Sprintf("OTA failed: %s", msg.ErrMsg)
		case 10:
			log.Printf("[TCP] OTA: UID=%s radar FW download complete", uid)
			progressMsg = "radar FW download complete"
		case 25:
			log.Printf("[TCP] OTA: UID=%s radar FW upgrade complete", uid)
			progressMsg = "radar FW upgrade complete"
		case 56:
			log.Printf("[TCP] OTA: UID=%s ESP FW download complete", uid)
			progressMsg = "ESP FW download complete"
		case 100:
			log.Printf("[TCP] OTA: UID=%s upgrade complete, device rebooting", uid)
			progressMsg = "upgrade complete, device rebooting"
		default:
			log.Printf("[TCP] OTA: UID=%s progress=%d", uid, msg.Progress)
			progressMsg = fmt.Sprintf("progress=%d", msg.Progress)
		}
		if onProgress != nil {
			onProgress(uid, int(msg.Progress), progressMsg)
		}

	default:
		uid := sm.GetUIDByConn(conn)
		log.Printf("[TCP] unhandled type=%d UID=%s len=%d", frame.Type, uid, len(frame.Data))
	}
}

func sendProto(conn net.Conn, msgType byte, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return WriteFrame(conn, &Frame{Type: msgType, Data: data})
}
