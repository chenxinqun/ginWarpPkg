package mux

import (
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/businessCodex"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/httpx/trace"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	WithoutTracePaths map[string]bool
)

func init() {
	// withoutLogPaths 这些请求，默认不记录链路追踪
	WithoutTracePaths = map[string]bool{
		"/metrics": true,

		"/debug/pprof/":             true,
		"/debug/pprof/cmdline":      true,
		"/debug/pprof/profile":      true,
		"/debug/pprof/symbol":       true,
		"/debug/pprof/trace":        true,
		"/debug/pprof/allocs":       true,
		"/debug/pprof/block":        true,
		"/debug/pprof/goroutine":    true,
		"/debug/pprof/heap":         true,
		"/debug/pprof/mutex":        true,
		"/debug/pprof/threadcreate": true,

		"/favicon.ico": true,

		"/system/health": true,
	}
}

func ErrorHandler(context Context, r Resource, err interface{}, opt Option) {
	logger := r.Logger
	stackInfo := string(debug.Stack())
	logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", stackInfo))
	context.AbortWithError(errno.New500Errno(businessCodex.GetServerErrorCode(), errno.Errorf("%+v", err)))

	if notify := opt.PanicNotify; notify != nil {
		notify(context, err, stackInfo)
	}
}

func AfterContext(ctx *gin.Context, ictx Context, r Resource, opt Option, ts time.Time) {
	logger := r.Logger
	// 非debug模式, 则捕获panic异常
	if gin.Mode() != gin.DebugMode {
		// 全局的异常处理
		if err := recover(); err != nil {
			ErrorHandler(ictx, r, err, opt)
		}
	}

	if ctx.Writer.Status() == http.StatusNotFound {
		return
	}

	var (
		response        interface{}
		businessCode    int
		businessCodeMsg string
		abortErr        error
		traceID         string
	)
	// 若上下文被终止, 在这里做处理
	if ctx.IsAborted() {
		for i := range ctx.Errors { // gin error
			multierr.AppendInto(&abortErr, ctx.Errors[i])
		}

		if err := ictx.abortError(); err != nil { // customer err
			multierr.AppendInto(&abortErr, err.GetErr())
			response = err
			businessCode = err.GetBusinessCode()
			businessCodeMsg = err.GetMsg()

			if x := ictx.Trace(); x != nil {
				ictx.SetHeader(trace.Header, x.ID())
				traceID = x.ID()
			}

			ctx.JSON(err.GetHttpCode(), &businessCodex.Response{
				Code: businessCode,
				Msg:  businessCodeMsg,
			})
		}
	} else {
		// 若上下文正常走完, 在这里处理
		response = ictx.getPayload()
		// 如果有payload
		if response != nil {
			// 记录链路ID
			if x := ictx.Trace(); x != nil {
				ictx.SetHeader(trace.Header, x.ID())
				traceID = x.ID()
			}
			// 返回json
			ctx.JSON(http.StatusOK, &businessCodex.Response{Data: response, Code: businessCodex.GetSucceedCode()})
		} else {
			fileResponse := ictx.getFilePayload()
			if fileResponse != nil {
				for k, v := range fileResponse {
					// 记录链路ID
					if x := ictx.Trace(); x != nil {
						ictx.SetHeader(trace.Header, x.ID())
						traceID = x.ID()
					}
					// 返回文件流
					contentType := "application/octet-stream"
					ContentDisposition := fmt.Sprintf(`attachment; filename=%s`, k)
					TransferEncoding := "binary"
					ctx.Header("content-Type", contentType)
					ctx.Header("Content-Disposition", ContentDisposition)
					ctx.Header("Content-Transfer-Encoding", TransferEncoding)
					ctx.Data(http.StatusOK, contentType, v)
				}
			}
		}
	}

	if opt.RecordMetrics != nil {
		uri := ictx.URI()
		if alias := ictx.Alias(); alias != "" {
			uri = alias
		}

		opt.RecordMetrics(
			ictx.Method(),
			uri,
			!ctx.IsAborted() && ctx.Writer.Status() == http.StatusOK,
			ctx.Writer.Status(),
			businessCode,
			time.Since(ts).Seconds(),
			traceID,
		)
	}

	var t *trace.Trace
	if x := ictx.Trace(); x != nil {
		t = x.(*trace.Trace)
	} else {
		return
	}

	decodedURL, _ := url.QueryUnescape(ctx.Request.URL.RequestURI())

	// ctx.Request.Header，精简 Header 参数
	traceHeader := map[string]string{
		"Content-Type":        ctx.GetHeader("Content-Type"),
		r.HeaderLoginToken:    ctx.GetHeader(r.HeaderLoginToken),
		r.HeaderSignToken:     ctx.GetHeader(r.HeaderSignToken),
		r.HeaderSignTokenDate: ctx.GetHeader(r.HeaderSignTokenDate),
	}

	t.WithRequest(&trace.Request{
		TTL:        "un-limit",
		Method:     ctx.Request.Method,
		DecodedURL: decodedURL,
		Header:     traceHeader,
		Body:       string(ictx.RawData()),
	})

	var responseBody interface{}

	if response != nil {
		responseBody = response
	}

	t.WithResponse(&trace.Response{
		Header:          ctx.Writer.Header(),
		HttpCode:        ctx.Writer.Status(),
		HttpCodeMsg:     http.StatusText(ctx.Writer.Status()),
		BusinessCode:    businessCode,
		BusinessCodeMsg: businessCodeMsg,
		Body:            responseBody,
		CostSeconds:     time.Since(ts).Seconds(),
	})

	t.Success = !ctx.IsAborted() && ctx.Writer.Status() == http.StatusOK
	t.CostSeconds = time.Since(ts).Seconds()

	logger.Info("mux-interceptor",
		zap.Any("method", ctx.Request.Method),
		zap.Any("path", decodedURL),
		zap.Any("http_code", ctx.Writer.Status()),
		zap.Any("business_code", businessCode),
		zap.Any("success", t.Success),
		zap.Any("cost_seconds", t.CostSeconds),
		zap.Any("trace_id", t.TraceID),
		zap.Any("trace_info", t),
		zap.Error(abortErr),
	)
}

func InitContext(r Resource, opt Option) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ts := time.Now()

		ictx := NewContext(ctx)
		defer ReleaseContext(ictx)

		ictx.init()

		// 不在这个列表中的URL, 开启链路追踪
		if !WithoutTracePaths[ctx.Request.URL.Path] {
			// 链路追踪原理, 从前端传过来一个链路ID, 然后就可以做全程链路追踪了. 如果是服务之间的互调,也请带上这个链路ID.
			if traceID := ictx.GetHeader(trace.Header); traceID != "" {
				ictx.setTrace(trace.New(traceID))
			} else {
				ictx.setTrace(trace.New(""))
			}
		}
		// 函数结束时执行这个匿名函数, 处理返回值
		defer AfterContext(ctx, ictx, r, opt, ts)
		ctx.Next()
	}
}
