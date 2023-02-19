package mux

import (
	"bytes"
	stdctx "context"
	"encoding/json"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/httpx/trace"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chenxinqun/ginWarpPkg/loggerx"

	"github.com/spf13/cast"

	"github.com/chenxinqun/ginWarpPkg/identify"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type HandlerFunc func(c Context)

type Trace = trace.T

const (
	// 考虑到一些值有可能会设置成 header, 并有可能会经过Nginx转发, 因此全部使用"-"做分割.
	_Alias            = "-alias-"
	_TraceName        = "-trace-"
	_LoggerName       = "-loggerx-"
	_BodyName         = "-body-"
	_PayloadName      = "-payload-"
	_GraphPayloadName = "-graph-payload-"
	_FilePayload      = "-file-payload-"
	UserID            = "-user-id-"
	TenantID          = "-tenant-id-"
	IsAdmin           = "-is-admin-"
	RoleType          = "-role-type-"
	UserName          = "-user-name-"
	_AbortErrorName   = "-abort-error-"
)

var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

func NewContext(ctx *gin.Context) Context {
	c := contextPool.Get().(*context)
	c.ctx = ctx
	return c
}

func ReleaseContext(ctx Context) {
	c := ctx.(*context)
	c.ctx = nil
	contextPool.Put(c)
}

var _ Context = (*context)(nil)

type Context interface {
	init()

	// ClientIP 返回可靠的客户端IP
	ClientIP() string

	// ShouldBindXML 反序列化 querystring
	// tag: `xml:"xxx"` (注：不要写成 query)
	ShouldBindXML(obj interface{}) *errno.Errno

	// ShouldBindQuery 反序列化 querystring
	// tag: `form:"xxx"` (注：不要写成 query)
	ShouldBindQuery(obj interface{}) *errno.Errno

	// ShouldBindPostForm 反序列化 postform (querystring会被忽略)
	// tag: `form:"xxx"`
	ShouldBindPostForm(obj interface{}) *errno.Errno

	// ShouldBindForm 同时反序列化 querystring 和 postform;
	// 当 querystring 和 postform 存在相同字段时，postform 优先使用。
	// tag: `form:"xxx"`
	ShouldBindForm(obj interface{}) *errno.Errno

	// ShouldBindJSON 反序列化 postjson
	// tag: `json:"xxx"`
	ShouldBindJSON(obj interface{}) *errno.Errno

	// ShouldBindURI 反序列化 path 参数(如路由路径为 /user/:name)
	// tag: `uri:"xxx"`
	ShouldBindURI(obj interface{}) *errno.Errno

	// Redirect 重定向
	Redirect(code int, location string)

	// Trace 获取 Trace 对象
	Trace() Trace
	setTrace(trace Trace)
	disableTrace()

	// Payload 正确返回
	Payload(payload interface{})
	getPayload() interface{}

	// FilePayload 文件流返回
	FilePayload(fileName string, FileContent []byte)
	getFilePayload() map[string][]byte

	// HTML 返回界面
	HTML(name string, obj interface{})

	// XML 返回XML响应
	XML(obj interface{})

	// String 返回String响应
	String(format string, values ...interface{})

	// AbortWithError 错误返回
	AbortWithError(err *errno.Errno)
	abortError() *errno.Errno

	// Header 获取 Header 对象
	Header() http.Header
	// GetHeader 获取 Header
	GetHeader(key string) string
	// SetHeader 设置 Header
	SetHeader(key, value string)

	// SetCookie 设置response cookie
	SetCookie(name, value string)
	// Cookie 获取request cookie
	Cookie(name string) (string, error)

	// IsAdmin 是否管理员
	IsAdmin() bool
	setIsAdmin(is int)
	// RoleType 角色类型
	RoleType() int32
	setRoleType(roleType int32)

	// TenantID 租户ID
	TenantID() int64
	setTenantID(tenantID int64)

	// UserID 获取 UserID
	UserID() int64
	setUserID(userID int64)

	// UserName 获取 UserName
	UserName() string
	setUserName(userName string)

	// Alias 设置路由别名 for metrics uri
	Alias() string
	setAlias(path string)

	// RequestInputParams 获取所有参数
	RequestInputParams() url.Values
	// RequestPostFormParams  获取 PostForm 参数
	RequestPostFormParams() url.Values
	FormFile(name string) (*multipart.FileHeader, error)
	// Request 获取 Request 对象
	Request() *http.Request
	// RawData 获取 Request.Body
	RawData() []byte
	// Method 获取 Request.Method
	Method() string
	// Host 获取 Request.Host
	Host() string
	// Path 获取 请求的路径 Request.URL.Path (不附带 querystring)
	Path() string
	// URI 获取 unescape 后的 Request.URL.RequestURI()
	URI() string

	// RequestContext (包装 Trace + Logger) 获取一个用来向别的服务发起请求的 context (当client关闭后，会自动canceled).
	// 可以传一个数字进来, 作为timeout的值, 单位是秒. 注意只能传0个或一个, 不要传多了
	RequestContext(timeout ...int) *StdContext

	// ResponseWriter 获取 ResponseWriter 对象
	ResponseWriter() gin.ResponseWriter

	// GinContext 获取gin的ctx
	GinContext() *gin.Context
}

type context struct {
	ctx *gin.Context
}

type StdContext struct {
	Cancel stdctx.CancelFunc
	stdctx.Context
	Trace
	*zap.Logger
}

func (c *context) init() {
	body, err := c.ctx.GetRawData()
	if err != nil {
		panic(err)
	}

	c.ctx.Set(_BodyName, body)                                   // redis body是为了trace使用
	c.ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // re-construct req body
}

func validateHeader(header string) (clientIP string, valid bool) {
	if header == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i, ipStr := range items {
		ipStr = strings.TrimSpace(ipStr)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return "", false
		}

		// We need to return the first IP in the list, but,
		// we should not early return since we need to validate that
		// the rest of the header is syntactically valid
		if i == 0 {
			clientIP = ipStr
			valid = true
			return
		}
	}
	return
}

// ClientIP 返回可靠的客户端IP
func (c *context) ClientIP() string {
	js, _ := json.Marshal(c.GinContext().Request.Header)
	loggerx.Default().Info("记录一下请求头", zap.String("url", c.Request().RequestURI), zap.ByteString("header", js))
	RemoteIPHeaders := []string{"-X-Client-IP-", "X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-IP"}
	for _, headerName := range RemoteIPHeaders {
		ip, valid := validateHeader(c.Header().Get(headerName))
		if valid {
			return ip
		}
	}
	return c.ctx.ClientIP()
}

// ShouldBindXML 反序列化xml请求
// tag: `xml:"xxx"` (注：不要写成query)
func (c *context) ShouldBindXML(obj interface{}) *errno.Errno {
	err := c.ctx.ShouldBindWith(obj, binding.XML)
	ret := errno.WrapParamBindErrno(err)
	return ret
}

// ShouldBindQuery 反序列化querystring
// tag: `form:"xxx"` (注：不要写成query)
func (c *context) ShouldBindQuery(obj interface{}) *errno.Errno {
	err := c.ctx.ShouldBindWith(obj, binding.Query)
	ret := errno.WrapParamBindErrno(err)
	return ret
}

// ShouldBindPostForm 反序列化 postform (querystring 会被忽略)
// tag: `form:"xxx"`
func (c *context) ShouldBindPostForm(obj interface{}) *errno.Errno {
	err := c.ctx.ShouldBindWith(obj, binding.FormPost)
	ret := errno.WrapParamBindErrno(err)
	return ret
}

// ShouldBindForm 同时反序列化querystring和postform;
// 当querystring和postform存在相同字段时，postform优先使用。
// tag: `form:"xxx"`
func (c *context) ShouldBindForm(obj interface{}) *errno.Errno {
	err := c.ctx.ShouldBindWith(obj, binding.Form)
	ret := errno.WrapParamBindErrno(err)
	return ret
}

// ShouldBindJSON 反序列化postjson
// tag: `json:"xxx"`
func (c *context) ShouldBindJSON(obj interface{}) *errno.Errno {
	err := c.ctx.ShouldBindWith(obj, binding.JSON)
	ret := errno.WrapParamBindErrno(err)
	return ret
}

// ShouldBindURI 反序列化path参数(如路由路径为 /user/:name)
// tag: `uri:"xxx"`
func (c *context) ShouldBindURI(obj interface{}) *errno.Errno {
	err := c.ctx.ShouldBindUri(obj)
	ret := errno.WrapParamBindErrno(err)
	return ret
}

func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	return c.GinContext().FormFile(name)
}

// Redirect 重定向
func (c *context) Redirect(code int, location string) {
	c.ctx.Redirect(code, location)
}

func (c *context) Trace() Trace {
	t, ok := c.ctx.Get(_TraceName)
	if !ok || t == nil {
		return nil
	}

	return t.(Trace)
}

// SetCookie 设置response cookie
func (c *context) SetCookie(name, value string) {
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	c.ctx.SetCookie(name, value, -1, "/", "", false, true)
}

// Cookie 获取request cookie
func (c *context) Cookie(name string) (string, error) {
	return c.ctx.Cookie(name)
}

func (c *context) setTrace(trace Trace) {
	c.ctx.Set(_TraceName, trace)
}

func (c *context) disableTrace() {
	c.setTrace(nil)
}

func (c *context) getPayload() interface{} {
	if payload, ok := c.ctx.Get(_PayloadName); ok {
		return payload
	}
	return nil
}

func (c *context) Payload(payload interface{}) {
	c.ctx.Set(_PayloadName, payload)
}

func (c *context) getGraphPayload() interface{} {
	if payload, ok := c.ctx.Get(_GraphPayloadName); ok {
		return payload
	}
	return nil
}

func (c *context) GraphPayload(payload interface{}) {
	c.ctx.Set(_GraphPayloadName, payload)
}

func (c *context) FilePayload(fileName string, fileContent []byte) {
	fileMap := make(map[string][]byte)
	fileMap[fileName] = fileContent
	c.ctx.Set(_FilePayload, fileMap)
}

func (c *context) getFilePayload() map[string][]byte {
	if payload, ok := c.ctx.Get(_FilePayload); ok {
		return payload.(map[string][]byte)
	}
	return nil
}

func (c *context) HTML(name string, obj interface{}) {
	c.ctx.HTML(http.StatusOK, name+".html", obj)
}

func (c *context) XML(obj interface{}) {
	c.ctx.XML(http.StatusOK, obj)
}

func (c *context) String(format string, values ...interface{}) {
	c.ctx.String(http.StatusOK, format, values...)
}

func (c *context) Header() http.Header {
	header := c.ctx.Request.Header

	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)

		clone[k] = value
	}
	return clone
}

func (c *context) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

func (c *context) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *context) IsAdmin() bool {
	c.setIsAdmin(-1)
	val, ok := c.ctx.Get(IsAdmin)
	if !ok {
		return false
	}

	return val.(bool)
}

func (c *context) setIsAdmin(arg int) {
	var is bool
	if arg == -1 {
		_, ok := c.ctx.Get(IsAdmin)
		if !ok {
			val := c.ctx.GetHeader(IsAdmin)
			if val != "" {
				is = cast.ToBool(val)
				c.ctx.Set(IsAdmin, is)
			}
		}
	} else {
		if arg == 1 {
			is = true
		}
		if arg == 0 {
			is = false
		}
		c.ctx.Set(IsAdmin, is)
	}
}

func (c *context) RoleType() int32 {
	c.setRoleType(0)
	val, ok := c.ctx.Get(RoleType)
	if !ok {
		return 0
	}
	return val.(int32)
}

func (c *context) setRoleType(roleType int32) {
	if roleType == 0 {
		_, ok := c.ctx.Get(RoleType)
		if !ok {
			val := c.ctx.GetHeader(RoleType)
			if identify.IsDigit(val) {
				num, _ := strconv.Atoi(val)
				c.ctx.Set(RoleType, int32(num))
			}
		}
	} else {
		c.ctx.Set(RoleType, roleType)
	}
}

func (c *context) TenantID() int64 {
	c.setTenantID(0)
	val, ok := c.ctx.Get(TenantID)
	if !ok {
		return 0
	}

	return val.(int64)
}

func (c *context) setTenantID(tenantID int64) {
	if tenantID == 0 {
		_, ok := c.ctx.Get(TenantID)
		if !ok {
			val := c.ctx.GetHeader(TenantID)
			if identify.IsDigit(val) {
				num, _ := strconv.ParseInt(val, 10, 64)
				c.ctx.Set(TenantID, num)
			}
		}
	} else {
		c.ctx.Set(TenantID, tenantID)
	}
}

func (c *context) UserID() int64 {
	c.setUserID(0)
	val, ok := c.ctx.Get(UserID)
	if !ok {
		return 0
	}

	return val.(int64)
}

func (c *context) setUserID(userID int64) {
	if userID == 0 {
		_, ok := c.ctx.Get(UserID)
		if !ok {
			val := c.ctx.GetHeader(UserID)
			if identify.IsDigit(val) {
				userID, _ = strconv.ParseInt(val, 10, 64)
				c.ctx.Set(UserID, userID)
			}
		}
	} else {
		c.ctx.Set(UserID, userID)
	}
}

func (c *context) UserName() string {
	c.setUserName("")
	val, ok := c.ctx.Get(UserName)
	if !ok {
		return ""
	}

	return val.(string)
}

func (c *context) setUserName(userName string) {
	if userName == "" {
		_, ok := c.ctx.Get(UserName)
		if !ok {
			val := c.ctx.GetHeader(UserName)
			if val != "" {
				c.ctx.Set(UserName, val)
			}
		}
	} else {
		c.ctx.Set(UserName, userName)
	}
}

func (c *context) AbortWithError(err *errno.Errno) {
	if err != nil {
		httpCode := err.GetHttpCode()
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}

		c.ctx.AbortWithStatus(httpCode)
		c.ctx.Set(_AbortErrorName, err)
	}
}

func (c *context) abortError() *errno.Errno {
	err, _ := c.ctx.Get(_AbortErrorName)
	return err.(*errno.Errno)
}

func (c *context) Alias() string {
	path, ok := c.ctx.Get(_Alias)
	if !ok {
		return ""
	}

	return path.(string)
}

func (c *context) setAlias(path string) {
	if path = strings.TrimSpace(path); path != "" {
		c.ctx.Set(_Alias, path)
	}
}

// RequestInputParams 获取所有参数
func (c *context) RequestInputParams() url.Values {
	_ = c.ctx.Request.ParseForm()
	return c.ctx.Request.Form
}

// RequestPostFormParams 获取 PostForm 参数
func (c *context) RequestPostFormParams() url.Values {
	_ = c.ctx.Request.ParseForm()
	return c.ctx.Request.PostForm
}

// Request 获取 Request
func (c *context) Request() *http.Request {
	return c.ctx.Request
}

func (c *context) RawData() []byte {
	body, ok := c.ctx.Get(_BodyName)
	if !ok {
		return nil
	}

	return body.([]byte)
}

// Method 请求的method
func (c *context) Method() string {
	return c.ctx.Request.Method
}

// Host 请求的host
func (c *context) Host() string {
	return c.ctx.Request.Host
}

// Path 请求的路径(不附带querystring)
func (c *context) Path() string {
	return c.ctx.Request.URL.Path
}

// URI unescape后的uri
func (c *context) URI() string {
	uri, _ := url.QueryUnescape(c.ctx.Request.URL.RequestURI())
	return uri
}

// GinContext 获取gin的context
func (c *context) GinContext() *gin.Context {
	return c.ctx
}

// RequestContext (包装 Trace ) 获取一个用来向别的服务发起请求的 context (当client关闭后，会自动canceled).
// 可以传一个数字进来, 作为timeout的值, 单位是秒. 注意只能传0个或一个, 不要传多了
func (c *context) RequestContext(timeout ...int) *StdContext {
	ret := GetRequestContext(timeout...)
	ret.Trace = c.Trace()
	return ret
}

// ResponseWriter 获取 ResponseWriter
func (c *context) ResponseWriter() gin.ResponseWriter {
	return c.ctx.Writer
}

func GetRequestContext(timeout ...int) *StdContext {
	var (
		tmout int
	)
	if len(timeout) < 1 || timeout[0] < 1 {
		tmout = 5
	} else {
		tmout = timeout[0]
	}
	ttl := time.Duration(tmout) * time.Second
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), ttl)
	ret := &StdContext{
		cancel,
		ctx,
		nil,
		nil,
	}
	runtime.SetFinalizer(ret, func(ctx *StdContext) {
		if ctx.Cancel != nil {
			ctx.Cancel()
		}
	})
	return ret
}
