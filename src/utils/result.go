package utils

// Result 是通用的返回结构体，包含状态码、消息和数据
type Result[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

// SuccessResult 返回成功结果（状态码 200，消息 "success"）
func SuccessResult[T any](data T) Result[T] {
	return Result[T]{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// ErrorResult 返回错误结果（状态码由调用者指定，消息由调用者传入）
func ErrorResult[T any](code int, message string) Result[T] {
	return Result[T]{
		Code:    code,
		Message: message,
	}
}
