package password

import "testing"

var basis = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func TestRandChar(t *testing.T) {
	s := RandChar(12, basis)
	t.Log(s)
	s1 := RandChar(12, basis)
	t.Log(s1)
	if s == s1 {
		t.Fatal("两次生成的随机字符串相等, 测试失败")
	}
}
