package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

var (
	salt = []byte(os.Getenv("SALT"))
)

type password string

func (p password) IsEqual(otherPassword string) bool {
	hashed := pbkdf2.Key([]byte(otherPassword), salt, 4096, 32, sha1.New)
	return p == password(base64.StdEncoding.EncodeToString(hashed))
}
