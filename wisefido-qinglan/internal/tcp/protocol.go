package tcp

// 清澜雷达 V2 TCP帧格式协议编解码
//
// 帧格式: [2字节Length(LE)][1字节type][protobuf data]
// - Length: 仅data长度，不包含Length和type自身（2+1=3字节头部）
// - type: 报文类型，决定protobuf消息类型
// - data: protobuf编码的消息体
//
// 参考: Java Demo ProcotolFrameDecoder
//   LengthFieldBasedFrameDecoder(LITTLE_ENDIAN, 1024, 0, 2, 1, 0)
//   lengthFieldOffset=0, lengthFieldLength=2, lengthAdjustment=1(type字节), initialBytesToStrip=0

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	// HeaderSize 帧头大小: 2字节length + 1字节type
	HeaderSize = 3
	// MaxFrameSize 最大帧大小 (与Java Demo一致)
	MaxFrameSize = 1024
)

// Frame 表示一个完整的TCP帧
type Frame struct {
	Type byte   // 报文类型 (1,2,3,4,5,7,8,...)
	Data []byte // protobuf编码数据
}

// ReadFrame 从连接中读取一个完整帧
// 阻塞读取，直到读完一个完整帧或连接断开
func ReadFrame(conn net.Conn) (*Frame, error) {
	// 1. 读取帧头 (3字节)
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("读取帧头失败: %w", err)
	}

	// 2. 解析length (Little-Endian, 2字节)
	dataLen := int(binary.LittleEndian.Uint16(header[0:2]))

	// 3. 解析type (1字节)
	msgType := header[2]

	// 4. 校验长度
	if dataLen < 0 || dataLen > MaxFrameSize {
		return nil, fmt.Errorf("帧长度异常: %d (最大%d)", dataLen, MaxFrameSize)
	}

	// 5. 读取data (protobuf)
	var data []byte
	if dataLen > 0 {
		data = make([]byte, dataLen)
		if _, err := io.ReadFull(conn, data); err != nil {
			return nil, fmt.Errorf("读取帧数据失败(期望%d字节): %w", dataLen, err)
		}
	}

	return &Frame{
		Type: msgType,
		Data: data,
	}, nil
}

// WriteFrame 将帧写入连接
// 参考Java Demo: ProtoBufCodecSharable.encode()
//   out.writeShortLE(bytes.readableBytes() - 1); // length = data长度(不含type)
//   out.writeBytes(bytes); // type + protobuf data
//
// 但注意Java的encode是: writeShortLE(总字节-1) 然后 writeBytes(type+data)
// 即 length = len(protobuf_data), 不包含type
func WriteFrame(conn net.Conn, f *Frame) error {
	// length = len(data), 不含type
	dataLen := len(f.Data)
	buf := make([]byte, HeaderSize+dataLen)

	// 2字节 length (LE)
	binary.LittleEndian.PutUint16(buf[0:2], uint16(dataLen))
	// 1字节 type
	buf[2] = f.Type
	// protobuf data
	if dataLen > 0 {
		copy(buf[3:], f.Data)
	}

	_, err := conn.Write(buf)
	return err
}
