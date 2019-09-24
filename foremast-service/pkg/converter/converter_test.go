package converter

import (
	"testing"
	"github.com/magiconair/properties/assert"
	"foremast.ai/foremast/foremast-service/pkg/models"
	"time"
	)

// TestConvertStatusToExternal
func TestConvertStatusToExternal(t *testing.T) {
	list := []string{"initial", "preprocess_inprogress", "postprocess_inprogress",
	"preprocess_completed","completed_unhealth", "completed_unknown", "preprocess_failed", "abort", "unknown"}
	rets := []string{"new", "inprogress", "inprogress",
		"inprogress","anomaly", "abort", "preprocess_failed", "abort", "unknown"}
	for i, v := range list {
		ret := ConvertStatusToExternal(v)
		assert.Equal(t, ret, rets[i])
	}
}

// TestConvertESToResp
func TestConvertESToResp(t *testing.T) {
	m := models.DocumentResponse{}
	m.StatusCode = "404"
	m.Status = "initial"
	m.ID = "uuid"
	m.Reason = "test"
	tm := time.Now()

	ld := models.LogDetail{"cpu", 100.0, 500.0, 50.0}
	lc := models.LogContent{50, "test", []models.LogDetail{ld}}
	hlog := models.HPALog{"uuid", &tm, &tm, float64(time.Now().Unix()), lc}
	ret := ConvertESToResp(m, []models.HPALog{hlog})
	assert.Equal(t, ret.JobID, "uuid")
}

// TestConvertESToHPAResp
func TestConvertESToHPAResp(t *testing.T) {
	tm := time.Now()
	ld := models.LogDetail{"cpu", 100.0, 500.0, 50.0}
	lc := models.LogContent{50, "test", []models.LogDetail{ld}}
	hlog := models.HPALog{"uuid", &tm, &tm, float64(time.Now().Unix()), lc}
	ret := ConvertESToHPAResp("uuid", []models.HPALog{hlog}, 200, "test")
	assert.Equal(t, ret.JobID, "uuid")
}

// TestConvertESToNewResp
func TestConvertESToNewResp(t *testing.T) {
	//func ConvertESToNewResp(uuid string, statusCode int32, status, reason string)
	ret := ConvertESToNewResp("uuid", 0, "initial", "test")
	assert.Equal(t, int32(200), ret.StatusCode)
}