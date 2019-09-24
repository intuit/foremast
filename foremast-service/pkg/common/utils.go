package common

import (
	"time"
	"log"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"github.com/gin-gonic/gin"
)

// StrToTime .... convert input string format of time to time
func StrToTime(input string) time.Time {
	t1, err := time.Parse(
		time.RFC3339,
		input)
	if err != nil {
		log.Print(err)
		return time.Now().UTC()
	}
	return t1
}

// UUIDGen .... based on input string to generate uuid
func UUIDGen(str string) string {
	secret := ""
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(str))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

// CheckStrEmpty .... check if input string is empty
func CheckStrEmpty(str string) bool {
	if len(strings.TrimSpace(str)) == 0 {
		return true
	}
	return false
}

// ErrorResponse .... use ErrorResponse to handle error
func ErrorResponse(c *gin.Context, code int, err string) {
	c.JSON(code, gin.H{
		"error": err,
	})
}