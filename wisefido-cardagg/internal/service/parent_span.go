package service

import "fmt"

// BuildParentSpan 北极星 datagram ref："<producer>.<seqN>"。
// 上游 envelope 缺 producer/seqN 时返回空串。
func BuildParentSpan(producer string, seqN uint64) string {
	if producer == "" || seqN == 0 {
		return ""
	}
	return fmt.Sprintf("%s.%d", producer, seqN)
}
