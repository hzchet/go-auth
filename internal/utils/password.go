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

func saltedHash(s string) string {
	return base64.StdEncoding.EncodeToString(pbkdf2.Key([]byte(s), salt, 4096, 32, sha1.New))
}

func AreEqual(p string, otherPassword string) bool {
	return p == saltedHash(otherPassword)
}
