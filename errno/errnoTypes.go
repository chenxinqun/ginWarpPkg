package errno

import (
	businessCodex2 "github.com/chenxinqun/ginWarpPkg/businessCodex"
	"github.com/chenxinqun/ginWarpPkg/httpx/validation"
	"net/http"
)

func NewBaseErrno(httpCode int, businessCode int, err error) *Errno {
	return NewErrno(httpCode, businessCode, businessCodex2.Text(businessCode)).WithErr(err)
}

// NewBusinessError 业务错误
// http code 203 表示业务错误, http code 401 表示请求错误
func NewBusinessErrno(businessCode int, err error) *Errno {
	if _, ok := businessCodex2.GetReturn401Map()[businessCode]; ok {
		return New401Errno(businessCode, err)
	}
	return New203Errno(businessCode, err)
}

func New500Errno(businessCode int, err error) *Errno {
	return NewBaseErrno(http.StatusInternalServerError, businessCode, err)
}

func New203Errno(businessCode int, err error) *Errno {
	return NewBaseErrno(http.StatusAccepted, businessCode, err)
}
func New400Errno(businessCode int, err error) *Errno {
	return NewBaseErrno(http.StatusBadRequest, businessCode, err)
}

func New401Errno(businessCode int, err error) *Errno {
	return NewBaseErrno(http.StatusUnauthorized, businessCode, err)
}

func New403Errno(businessCode int, err error) *Errno {
	return NewBaseErrno(http.StatusForbidden, businessCode, err)
}

func New404Errno(businessCode int, err error) *Errno {
	return NewBaseErrno(http.StatusNotFound, businessCode, err)
}

func New429Errno(businessCode int, err error) *Errno {
	return NewBaseErrno(http.StatusTooManyRequests, businessCode, err)
}

// WrapParamBindError 请求参数绑定到go对象错误.
// 请求参数序列化错误.
func WrapParamBindErrno(err error) *Errno {
	if err != nil {
		ret := NewErrno(http.StatusBadRequest, businessCodex2.GetParamBindErrorCode(),
			businessCodex2.Text(businessCodex2.GetParamBindErrorCode())+": "+validation.Error(err)).WithErr(err)
		return ret
	}

	return nil
}

// WrapMySQLExecError SQL创建, 更新, 删除 执行错误. 查询错误不要用这个. 唯一索引和主键冲突, 一般是业务问题, 也不要用这个.
func WrapMySQLExecErrno(err error) *Errno {
	if err != nil {
		return New500Errno(businessCodex2.GetMySQLExecErrorCode(), err)
	}

	return nil
}
