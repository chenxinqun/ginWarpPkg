package ras

import (
	"log"
	"testing"
)

func TestEncryptRSA(t *testing.T) {
	pwd := "123456"
	pub := "-----BEGIN RSA Public Key-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDVe7UNPpG0oEmAJBlfSlQ+625o\nDpAK1rgCWWfyqJavt+LzQEQO9bco0bugOq+xlurUsLJHTSf0X0730TNIOaiONlLf\nPe2xzRZlMNtoYl7cIwE+AanNLHj0g9JBzXLRVSn9ikAwUuSDBRJ1K7hHm34DUVkz\nRbyY7mGS6Hst84mq5QIDAQAB\n-----END RSA Public Key-----\n"
	pem := "-----BEGIN RSA Private Key-----\nMIICXQIBAAKBgQDVe7UNPpG0oEmAJBlfSlQ+625oDpAK1rgCWWfyqJavt+LzQEQO\n9bco0bugOq+xlurUsLJHTSf0X0730TNIOaiONlLfPe2xzRZlMNtoYl7cIwE+AanN\nLHj0g9JBzXLRVSn9ikAwUuSDBRJ1K7hHm34DUVkzRbyY7mGS6Hst84mq5QIDAQAB\nAoGAZiTLiuu+EXuDz3D2RtasmnJRIC6fkuALqOwYRU2O08KbLyI3riS5HynCqTaL\nK+B2uY9VrbHoBQ+5G++Xpt4XnBtU6fOARAuGZkN03MgyNkyFiDaCJG4MUaeWPEE5\nazAEimWYumh11AC3CLpUwVGLa72VTCWWqCuo0VlUF8pIPMECQQD8rHFU85IqGD7b\n/SGljc8XEz9Mkc4Q4vVMeZyNUa/qSFfZ6lEHz6EaWHnwFYjqnWa8YsjaFCXR2V6V\n+frwGyd1AkEA2EsvhnOZojzw7oGlJWgACQhmTw58ZEOqOq+vBvvXi6/Rk93Ln+NE\nC1dwDraUBe67rb+QTXolKJcnFlDTijx3sQJAJkXcmNiYMEYh52KtYQ1c7ArfULLZ\nOteV/nKBUyqncd5paDnE8mDx7zKtrb8lURxsfmacM+RPYj0BxcfqycnjLQJBAIdG\nHqccTY3mR1kjxEGs1bjQhAwVpz6eAy1JC1J218wJXi34nY2V+cyOFwtcrR84vDBi\nisGqDutf/ZY7XtIqF0ECQQDA9Ey+zUW8O+zADJ9+sdBZrSOtZRn9Bi4MbNcg1B/T\noaV9RxkVDU2kxjEfeNcZX8tFnx2CnxXxPIE4eP3paTvz\n-----END RSA Private Key-----\n"
	cipher, err := EncryptRSA(pwd, pub)
	if err != nil {
		log.Fatal(err)
	}
	plain, err := DecryptRSA(cipher, pem)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(pwd, cipher, plain)
}

func TestGenerateRSAKey(t *testing.T) {
	pem, pub, err := GenerateRSAKey(1024)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pem, pub)
}
