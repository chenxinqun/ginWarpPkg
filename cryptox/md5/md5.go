package md5

import (
	cryptoMD5 "crypto/md5"
	"encoding/hex"
)

func SumByte(byt []byte) string {
	s := cryptoMD5.New()
	s.Write(byt)
	return hex.EncodeToString(s.Sum(nil))
}

func SumString(str string) string {
	return SumByte([]byte(str))
}
