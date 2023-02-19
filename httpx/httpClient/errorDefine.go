package httpClient

var _ ReplyErr = (*replyError)(nil)

// ReplyErr 错误响应，当 resp.StatusCode != http.StatusOK 时用来包装返回的 httpcode 和 body 。
type ReplyErr interface {
	error
	StatusCode() int
	Body() []byte
}

type replyError struct {
	err        error
	statusCode int
	body       []byte
}

func (r *replyError) Error() string {
	return r.err.Error()
}

func (r *replyError) StatusCode() int {
	return r.statusCode
}

func (r *replyError) Body() []byte {
	return r.body
}

func newReplyErr(statusCode int, body []byte, err error) ReplyErr {
	return &replyError{
		statusCode: statusCode,
		body:       body,
		err:        err,
	}
}

// ToReplyErr 尝试将 err 转换为 ReplyErr
func ToReplyErr(err error) (ReplyErr, bool) {
	if err == nil {
		return nil, false
	}
	e, ok := err.(ReplyErr)

	return e, ok
}
