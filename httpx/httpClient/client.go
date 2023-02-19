package httpClient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/httpx/trace"
	"net/http"
	httpURL "net/url"
	"time"
)

const (
	// DefaultTTL 一次http请求最长执行1分钟.
	DefaultTTL = time.Minute
)

type RequestFunc func(url string, requestData interface{}, options ...OptionHandler) (body []byte, httpCode int, err error)

// GetJson get 请求.
func GetJson(url string, requestData interface{}, options ...OptionHandler) (body []byte, httpCode int, err error) {
	form := requestData.(httpURL.Values)
	return withOutJsonBody(http.MethodGet, url, form, options...)
}

// DeleteJson delete 请求.
func DeleteJson(url string, requestData interface{}, options ...OptionHandler) (body []byte, httpCode int, err error) {
	form := requestData.(httpURL.Values)
	return withOutJsonBody(http.MethodDelete, url, form, options...)
}

// PostJSON post json 请求.
func PostJSON(url string, requestData interface{}, options ...OptionHandler) (body []byte, httpCode int, err error) {
	raw := requestData.(json.RawMessage)
	return withJSONBody(http.MethodPost, url, raw, options...)
}

// PutJSON put json 请求.
func PutJSON(url string, requestData interface{}, options ...OptionHandler) (body []byte, httpCode int, err error) {
	raw := requestData.(json.RawMessage)
	return withJSONBody(http.MethodPut, url, raw, options...)
}

// PatchJSON patch json 请求.
func PatchJSON(url string, requestData interface{}, options ...OptionHandler) (body []byte, httpCode int, err error) {
	raw := requestData.(json.RawMessage)
	return withJSONBody(http.MethodPatch, url, raw, options...)
}

func withOutJsonBody(method, url string, form httpURL.Values, options ...OptionHandler) (body []byte, httpCode int, err error) {
	if url == "" {
		return nil, 0, errno.NewError("url required")
	}

	if len(form) > 0 {
		if url, err = addFormValuesIntoURL(url, form); err != nil {
			return
		}
	}

	ts := time.Now()

	opt := getOption()
	defer func() {
		if opt.trace != nil {
			opt.dialog.Success = err == nil
			opt.dialog.CostSeconds = time.Since(ts).Seconds()
			opt.trace.AppendDialog(opt.dialog)
		}

		releaseOption(opt)
	}()

	for _, f := range options {
		f(opt)
	}
	opt.header["Content-Type"] = []string{"application/json; charset=utf-8"}
	if opt.trace != nil {
		opt.header[trace.Header] = []string{opt.trace.ID()}
	}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()

	if opt.dialog != nil {
		decodedURL, _ := httpURL.QueryUnescape(url)
		opt.dialog.Request = &trace.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: decodedURL,
			Header:     opt.header,
		}
	}

	body, httpCode, err = doHTTP(ctx, method, url, nil, opt)
	return
}

func withFormBody(method, url string, form httpURL.Values, options ...OptionHandler) (body []byte, httpCode int, err error) {
	if url == "" {
		return nil, 0, errno.NewError("url required")
	}
	if len(form) == 0 {
		return nil, 0, errno.NewError("form required")
	}

	ts := time.Now()

	opt := getOption()
	defer func() {
		if opt.trace != nil {
			opt.dialog.Success = err == nil
			opt.dialog.CostSeconds = time.Since(ts).Seconds()
			opt.trace.AppendDialog(opt.dialog)
		}

		releaseOption(opt)
	}()

	for _, f := range options {
		f(opt)
	}
	opt.header["Content-Type"] = []string{"application/x-www-form-urlencoded; charset=utf-8"}
	if opt.trace != nil {
		opt.header[trace.Header] = []string{opt.trace.ID()}
	}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()

	formValue := form.Encode()
	if opt.dialog != nil {
		decodedURL, _ := httpURL.QueryUnescape(url)
		opt.dialog.Request = &trace.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: decodedURL,
			Header:     opt.header,
			Body:       formValue,
		}
	}
	body, httpCode, err = doHTTP(ctx, method, url, []byte(formValue), opt)
	return
}

func withJSONBody(method, url string, raw json.RawMessage, options ...OptionHandler) (body []byte, httpCode int, err error) {
	if url == "" {
		return nil, 0, errno.NewError("url required")
	}
	if len(raw) == 0 {
		return nil, 0, errno.NewError("raw required")
	}

	ts := time.Now()

	opt := getOption()
	defer func() {
		if opt.trace != nil {
			opt.dialog.Success = err == nil
			opt.dialog.CostSeconds = time.Since(ts).Seconds()
			opt.trace.AppendDialog(opt.dialog)
		}

		releaseOption(opt)
	}()

	for _, f := range options {
		f(opt)
	}
	opt.header["Content-Type"] = []string{"application/json; charset=utf-8"}
	if opt.trace != nil {
		opt.header[trace.Header] = []string{opt.trace.ID()}
	}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()

	if opt.dialog != nil {
		decodedURL, _ := httpURL.QueryUnescape(url)
		opt.dialog.Request = &trace.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: decodedURL,
			Header:     opt.header,
			Body:       string(raw), // TODO unsafe
		}
	}

	defer func() {
		if opt.alarmObject == nil {
			return
		}

		if opt.alarmVerify != nil && !opt.alarmVerify(body) && err == nil {
			return
		}

		info := &struct {
			TraceID string `json:"trace_id"`
			Request struct {
				Method string `json:"method"`
				URL    string `json:"url"`
			} `json:"request"`
			Response struct {
				HTTPCode int    `json:"http_code"`
				Body     string `json:"body"`
			} `json:"response"`
			Error string `json:"error"`
		}{}

		if opt.trace != nil {
			info.TraceID = opt.trace.ID()
		}
		info.Request.Method = method
		info.Request.URL = url
		info.Response.HTTPCode = httpCode
		info.Response.Body = string(body)
		info.Error = ""
		if err != nil {
			info.Error = fmt.Sprintf("%+v", err)
		}

		raw, _ := json.MarshalIndent(info, "", " ")
		onFailedAlarm(opt.alarmTitle, raw, opt.logger, opt.alarmObject)
	}()
	body, httpCode, err = doHTTP(ctx, method, url, raw, opt)
	return
}
