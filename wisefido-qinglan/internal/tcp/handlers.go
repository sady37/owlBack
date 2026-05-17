package tcp

// OTA-focused message handler (simplified from ota-main)

import (
	"fmt"
	"net"

	pb "wisefido-qinglan/proto/gen"

	"go.uber.org/zap"
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
	lg := sm.logger
	switch frame.Type {

	case TypeGetServer:
		req := &pb.GetServerReq{}
		if err := proto.Unmarshal(frame.Data, req); err != nil {
			lg.Warn("tcp type 1 decode", zap.Error(err))
			return
		}
		lg.Info("tcp dispatch request", zap.String("uid", req.Uid), zap.String("type", req.Type))
		resp := &pb.GetServerResponse{Seq: 2, Result: 0, Server: serverAddr, Port: serverPort}
		sendProto(conn, TypeGetServerResp, resp)

	case TypeRegister:
		req := &pb.RegisterReq{}
		if err := proto.Unmarshal(frame.Data, req); err != nil {
			lg.Warn("tcp type 3 decode", zap.Error(err))
			return
		}
		lg.Info("tcp register",
			zap.String("uid", req.Uid),
			zap.String("type", req.Type),
			zap.String("sf_ver", req.Sfver),
			zap.String("hw_ver", req.Hwver),
		)
		sm.Connect(conn, req.Uid, req.Type, req.Sfver, req.Hwver)
		resp := &pb.RegisterResponse{Seq: 4, Result: 0}
		sendProto(conn, TypeRegisterResp, resp)
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
			lg.Warn("tcp type 17 decode", zap.Error(err))
			return
		}
		uid := sm.GetUIDByConn(conn)
		if msg.Result == 0 {
			lg.Info("tcp ota response: accepted", zap.String("uid", uid))
			if onProgress != nil {
				onProgress(uid, 0, "device accepted OTA")
			}
		} else {
			lg.Info("tcp ota response: rejected",
				zap.String("uid", uid),
				zap.Int32("result", msg.Result),
				zap.String("err_msg", msg.Errmsg),
			)
			if onProgress != nil {
				onProgress(uid, -1, fmt.Sprintf("device rejected: %s", msg.Errmsg))
			}
		}

	case TypeOTAProgress:
		msg := &pb.OTAProgress{}
		if err := proto.Unmarshal(frame.Data, msg); err != nil {
			lg.Warn("tcp type 18 decode", zap.Error(err))
			return
		}
		uid := sm.GetUIDByConn(conn)
		progressMsg := ""
		switch msg.Progress {
		case -1:
			lg.Warn("tcp ota failed", zap.String("uid", uid), zap.String("err_msg", msg.ErrMsg))
			progressMsg = fmt.Sprintf("OTA failed: %s", msg.ErrMsg)
		case 10:
			lg.Info("tcp ota: radar fw download complete", zap.String("uid", uid))
			progressMsg = "radar FW download complete"
		case 25:
			lg.Info("tcp ota: radar fw upgrade complete", zap.String("uid", uid))
			progressMsg = "radar FW upgrade complete"
		case 56:
			lg.Info("tcp ota: esp fw download complete", zap.String("uid", uid))
			progressMsg = "ESP FW download complete"
		case 100:
			lg.Info("tcp ota: upgrade complete, device rebooting", zap.String("uid", uid))
			progressMsg = "upgrade complete, device rebooting"
		default:
			lg.Info("tcp ota progress", zap.String("uid", uid), zap.Int32("progress", msg.Progress))
			progressMsg = fmt.Sprintf("progress=%d", msg.Progress)
		}
		if onProgress != nil {
			onProgress(uid, int(msg.Progress), progressMsg)
		}

	default:
		uid := sm.GetUIDByConn(conn)
		lg.Warn("tcp unhandled frame",
			zap.Uint8("type", frame.Type),
			zap.String("uid", uid),
			zap.Int("len", len(frame.Data)),
		)
	}
}

func sendProto(conn net.Conn, msgType byte, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return WriteFrame(conn, &Frame{Type: msgType, Data: data})
}
