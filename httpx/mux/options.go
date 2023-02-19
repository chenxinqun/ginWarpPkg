package mux

import (
	"fmt"
)

const MaxBurstSize = 100000

type OptionHandler func(*Option)

type Option struct {
	DisablePProf      bool
	DisableSwagger    bool
	DisablePrometheus bool
	PanicNotify       OnPanicNotify
	RecordMetrics     RecordMetrics
	EnableCors        bool
	EnableRate        bool
}

// OnPanicNotify 发生panic时通知用
type OnPanicNotify func(ctx Context, err interface{}, stackInfo string)

// RecordMetrics 记录prometheus指标用
// 如果使用AliasForRecordMetrics配置了别名，uri将被替换为别名。
type RecordMetrics func(method, uri string, success bool, httpCode, businessCode int, costSeconds float64, traceID string)

// WithDisablePProf 禁用 pprof
func WithDisablePProf() OptionHandler {
	return func(opt *Option) {
		opt.DisablePProf = true
	}
}

// WithDisableSwagger 禁用 swagger
func WithDisableSwagger() OptionHandler {
	return func(opt *Option) {
		opt.DisableSwagger = true
	}
}

// WithDisablePrometheus 禁用prometheus
func WithDisablePrometheus() OptionHandler {
	return func(opt *Option) {
		opt.DisablePrometheus = true
	}
}

// WithRecordMetrics 设置记录prometheus记录指标回调
func WithRecordMetrics(record RecordMetrics) OptionHandler {
	return func(opt *Option) {
		opt.RecordMetrics = record
	}
}

// WithEnableCors 开启CORS
func WithEnableCors() OptionHandler {
	return func(opt *Option) {
		opt.EnableCors = true
		fmt.Println("* [register cors]")
	}
}

func WithEnableRate() OptionHandler {
	return func(opt *Option) {
		opt.EnableRate = true
		fmt.Println("* [register rate]")
	}
}

func DisableTrace(ctx Context) {
	ctx.disableTrace()
}

// AliasForRecordMetrics 对请求uri起个别名，用于prometheus记录指标。
// 如：Get /user/:username 这样的uri，因为username会有非常多的情况，这样记录prometheus指标会非常的不有好。
func AliasForRecordMetrics(path string) HandlerFunc {
	return func(ctx Context) {
		ctx.setAlias(path)
	}
}
