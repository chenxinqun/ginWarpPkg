package identify

import (
	"regexp"
)

var (
	digitM *regexp.Regexp
	floatM *regexp.Regexp
)

func init() {
	digitM, _ = regexp.Compile(`^\d+$`)
	floatM, _ = regexp.Compile(`^\d+.\d+$`)
}

func IsDigit(s string) bool {
	return digitM.Match([]byte(s))
}

func IsFloat(s string) bool {
	return floatM.Match([]byte(s))
}
