package common

import (
	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties/assert"
	"net/http/httptest"
	"testing"
	assert2 "github.com/stretchr/testify/assert"
	"time"
)

// TestErrorResponse
func TestErrorResponse(t *testing.T) {
	resp := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(resp)
	ErrorResponse(c, 200, "test_error")
	assert.Equal(t, resp.Body.String(), "{\"error\":\"test_error\"}")
}

// TestCheckStrEmpty
func TestCheckStrEmpty(t *testing.T) {
	assert2.True(t, CheckStrEmpty(""))
	assert2.False(t, CheckStrEmpty("abc"))
}

// TestUUIDGen
func TestUUIDGen(t *testing.T) {
	ret := UUIDGen("abc")
	assert.Equal(t, ret, "fd7adb152c05ef80dccf50a1fa4c05d5a3ec6da95575fc312ae7c5d091836351")
}

// TestStrToTime
func TestStrToTime(t *testing.T) {
	ts := StrToTime("2006-01-02T15:04:05Z")
	t1 := time.Date(2006,  time.January, 2, 15, 4, 5, 0, time.UTC)
	assert.Equal(t, ts, t1)
	ts2 := StrToTime("wrong_string")
	assert2.Equal(t, ts2.YearDay(), time.Now().YearDay())
}