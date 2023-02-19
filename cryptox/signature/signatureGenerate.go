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

	"github.com/chenxinqun/ginWarpPkg/timex"
)

// Generate
// path 请求的路径 (不附带 querystring)
func (s *signature) Generate(path string, method string, params url.Values) (authorization, date string, err error) {
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

	// Date
	date = timex.CSTLayoutString()

	// Encode() 方法中自带 sorted by key
	sortParamsEncode, err := url.QueryUnescape(params.Encode())
	if err != nil {
		err = errno.Errorf("url QueryUnescape error %v", err)

		return
	}

	// 加密字符串规则
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(path)
	buffer.WriteString(delimiter)
	buffer.WriteString(methodName)
	buffer.WriteString(delimiter)
	buffer.WriteString(sortParamsEncode)
	buffer.WriteString(delimiter)
	buffer.WriteString(date)

	// 对数据进行 sha256 加密，并进行 base64 encode
	hash := hmac.New(sha256.New, []byte(s.secret))
	hash.Write(buffer.Bytes())
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	authorization = fmt.Sprintf("%s %s", s.key, digest)

	return
}
