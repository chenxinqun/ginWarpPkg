package signature

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"net/url"
	"strings"
	"time"

	"github.com/chenxinqun/ginWarpPkg/timex"
)

func (s *signature) Verify(authorization, date string, path string, method string,
	params url.Values) (ok bool, err error) {
	if date == "" {
		err = errno.NewError("date required")

		return
	}

	if path == "" {
		err = errno.NewError("path required")

		return
	}

	if method == "" {
		err = errno.NewError("method required")
		return
	}

	methodName := strings.ToUpper(method)
	if !methods[methodName] {
		err = errno.NewError("method param error")
		return
	}

	ts, err := timex.ParseCSTInLocation(date)
	if err != nil {
		err = errno.NewError("date must follow '2006-01-02 15:04:05'")
		return
	}

	if timex.SubInLocation(ts) > float64(s.ttl/time.Second) {
		err = errno.Errorf("date exceeds limit %v", s.ttl)
		return
	}

	// Encode() 方法中自带 sorted by key
	sortParamsEncode, err := url.QueryUnescape(params.Encode())
	if err != nil {
		err = errno.Errorf("url QueryUnescape error %v", err)
		return
	}

	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(path)
	buffer.WriteString(delimiter)
	buffer.WriteString(methodName)
	buffer.WriteString(delimiter)
	buffer.WriteString(sortParamsEncode)
	buffer.WriteString(delimiter)
	buffer.WriteString(date)

	// 对数据进行 hmac 加密，并进行 base64 encode
	hash := hmac.New(sha256.New, []byte(s.secret))
	hash.Write(buffer.Bytes())
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	ok = authorization == fmt.Sprintf("%s %s", s.key, digest)
	return
}
