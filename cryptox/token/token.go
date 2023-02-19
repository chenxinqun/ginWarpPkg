package token

import (
	"github.com/dgrijalva/jwt-go"
	"net/url"
	"time"
)

var _ Token = (*token)(nil)

type Token interface {

	// JwtSign 签名
	JwtSign(userID int64, userName string, expireDuration time.Duration) (tokenString string, err error)

	// JwtParse 解密
	JwtParse(tokenString string) (*Claims, error)

	// UrlSign URL 签名方式，不支持解密
	UrlSign(path string, method string, params url.Values) (tokenString string, err error)
}

type token struct {
	secret string
}

type Claims struct {
	UserID   int64
	UserName string
	jwt.StandardClaims
}

func New(secret string) Token {
	return &token{
		secret: secret,
	}
}

func (t *token) i() {}
