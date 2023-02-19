package idGen

import gouuid "github.com/satori/go.uuid"

func NewUUID4() (uuid4 string) {
	uuid4 = gouuid.NewV4().String()

	return
}
