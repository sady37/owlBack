package httpapi

// Result 与 owlFront 的 `types/axios.d.ts` 保持一致
// - code: ResultEnum.SUCCESS = 2000
// - type: 'success' | 'error' | 'warning'
// - message: string
// - result: any
type Result[T any] struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Result  T      `json:"result"`
}

const (
	ResultSuccess = 2000
	ResultError   = -1
	// TokenExpired 使用 code=60401 + HTTP 401（前端 Axios 拦截器会特殊处理）
	ResultTokenExpired = 60401
)

func Ok[T any](result T) Result[T] {
	return Result[T]{Code: ResultSuccess, Type: "success", Message: "ok", Result: result}
}

func Fail(message string) Result[any] {
	return Result[any]{Code: ResultError, Type: "error", Message: message, Result: nil}
}


