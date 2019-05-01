package converter

import (
	"strconv"

	"foremast.ai/foremast/foremast-service/pkg/models"
)

// ConvertStatusToExternal .... convert internal status to k8s controller
func ConvertStatusToExternal(status string) string {
	switch status {
	case "initial":
		return "new"
	case "preprocess_inprogress", "postprocess_inprogress", "preprocess_completed":
		return "inprogress"
	case "completed_health":
		return "success"
	case "completed_unhealth":
		return "anomaly"
	case "completed_unknown":
		return "abort"
	case "preprocess_failed":
		return "abort"
	case "abort":
		return "abort"
	default:
		return "unknown"
	}
}

// ConvertESToNewResp .... convert elasticsearch to new response
func ConvertESToNewResp(uuid string, statusCode int32, status, reason string) models.ApplicationHealthAnalyzeResponseNew {
	if statusCode == 0 {
		statusCode = 200
	}

	resp := models.ApplicationHealthAnalyzeResponseNew{
		JobID:      uuid,
		StatusCode: statusCode,
		Status:     status,
		Reason:     reason,
	}
	return resp
}

// ConvertESToHPAResp .... convert elasticsearch to new HPA logs response
func ConvertESToHPAResp(jobID string, logs []models.HPALog, statusCode int32, reason string) models.HPALogResponse {
	resp := models.HPALogResponse{
		JobID: jobID, StatusCode: statusCode}
	if reason != "" {
		resp.Reason = reason
	}
	hlogs := make([]models.HPALog, 0)
	for _, log := range logs {
		hlogs = append(hlogs, models.HPALog{Timestamp: log.Timestamp, Log: log.Log})
	}
	resp.HPALog = hlogs
	return resp
}

// ConvertESToResp .... convert elasticsearch to response
func ConvertESToResp(input models.DocumentResponse, logs []models.HPALog) models.ApplicationHealthAnalyzeResponse {

	code, err := strconv.Atoi(input.StatusCode)
	if err != nil {
		code = 200
	}
	resp := models.ApplicationHealthAnalyzeResponse{
		JobID:      input.ID,
		StatusCode: int32(code),
		Status:     ConvertStatusToExternal(input.Status),
		Reason:     input.Reason,
	}
	if logs != nil {
		hpaLogs := make([]map[string]interface{}, 0)
		for _, l := range logs {
			hpaLogEntry := map[string]interface{}{}
			hpaLogEntry["timestamp"] = l.Timestamp
			hpaLog := map[string]interface{}{}
			hpaLog["hpascore"] = l.Log.HPAScore
			hpaLog["reason"] = l.Log.Reason
			details := []map[string]interface{}{}
			for _, d := range l.Log.Details {
				detail := map[string]interface{}{}
				detail["metricAlias"] = d.MetricType
				detail["current"] = d.Current
				detail["upper"] = d.Upper
				detail["lower"] = d.Lower
				details = append(details, detail)
			}
			hpaLog["details"] = details
			hpaLogEntry["hpalog"] = hpaLog
			hpaLogs = append(hpaLogs, hpaLogEntry)
		}
		resp.HPALog = hpaLogs
	}
	return resp
}
