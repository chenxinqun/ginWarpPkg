package errno

import (
	"encoding/json"
)

var _ error = (*Errno)(nil)

type Errno struct {
	HttpCode     int    `json:"http_code"`     // HTTP Code
	BusinessCode int    `json:"business_code"` // Business Code
	Message      string `json:"message"`       // 描述信息
	err          Error  // 错误信息
}

func NewErrno(httpCode, businessCode int, msg string) *Errno {
	return &Errno{
		HttpCode:     httpCode,
		BusinessCode: businessCode,
		Message:      msg,
	}
}

func (e *Errno) Error() string {
	if e.err != nil {
		return e.err.Error()
	}

	return ""
}

func (e *Errno) WithErr(err error) *Errno {
	e.err = WithStack(err)
	return e
}

func (e *Errno) GetHttpCode() int {
	return e.HttpCode
}

func (e *Errno) GetBusinessCode() int {
	return e.BusinessCode
}

func (e *Errno) GetMsg() string {
	return e.Message
}

func (e *Errno) GetErr() error {
	return e.err
}

// ToString 返回 JSON 格式的错误详情
func (e *Errno) ToString() string {
	raw, _ := json.Marshal(e)

	return string(raw)
}
