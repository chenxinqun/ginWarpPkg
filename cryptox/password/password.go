package password

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5" //nolint:gosec
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

const (
	saltPassword    = "!@!$!dsdfdAASDFASDGASADSFAHGFKGHJKLMJLI&*)%&*%^&#$%^"
	defaultPassword = "123456"
)

func GeneratePassword(str string) (password string) {
	var m = md5.New() //nolint:gosec
	m.Write([]byte(str))
	mByte := m.Sum(nil)

	// hmac
	h := hmac.New(sha256.New, []byte(saltPassword))
	h.Write(mByte)
	password = hex.EncodeToString(h.Sum(nil))

	return
}

func ResetPassword() (password string) {
	var m = md5.New() //nolint:gosec
	m.Write([]byte(defaultPassword))
	mStr := hex.EncodeToString(m.Sum(nil))
	password = GeneratePassword(mStr)

	return
}

func GenerateLoginToken(id int32) (token string) {
	m := md5.New() //nolint:gosec
	m.Write([]byte(fmt.Sprintf("%d%s", id, saltPassword)))
	token = hex.EncodeToString(m.Sum(nil))

	return
}

func RandChar(size int, char string) string {
	var s bytes.Buffer
	// 休眠1纳秒, 避免连续两次创建的随机字符串相等
	time.Sleep(time.Nanosecond)
	rand.Seed(time.Now().UnixNano()) // 产生随机种子
	for i := 0; i < size; i++ {
		rn := rand.Int63()
		l := int64(len(char))
		confirm := rn % l
		s.WriteByte(char[confirm])
	}
	return s.String()
}
