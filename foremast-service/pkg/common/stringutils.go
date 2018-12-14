package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func UUIDGen(str string) string {
	secret := ""
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(str))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

func CheckStrEmpty(str string) bool {
	if len(strings.TrimSpace(str))==0{
		return true
	}
	return false
}


