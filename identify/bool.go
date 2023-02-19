package identify

import (
	"regexp"
	"strings"
)

var (
	boolM *regexp.Regexp
)

func init() {
	boolM, _ = regexp.Compile(`^(true)|(false)$`)
}

func IsBool(s string) bool {
	// 转换为小写再进行比对
	s = strings.ToLower(s)
	return boolM.Match([]byte(s))
}
