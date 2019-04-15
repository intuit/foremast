package converter

import (
	"fmt"
	"strconv"

	models "foremast.ai/foremast/foremast-service/pkg/models"
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
		return "inprogress"
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

// ConvertESToNewResp .... convert elasticsearch to new response
// func ConvertESToHPAResp(appName string, namespace string, mtime time.Time, ctime time.Time, status string, id string) models.HPAAlertResponse {
func ConvertESToHPAResp(doc models.DocumentResponse, logs []models.HPALog) models.HPAAlertResponse {
	// if statusCode == 0 {
	// 	statusCode = 200
	// }

	resp := models.HPAAlertResponse{
		JobID:      doc.ID,
		AppName:    doc.AppName,
		Namespace:  doc.Namespace,
		Strategy:   "hpa",
		ModifiedAt: doc.ModifiedAt,
		CreatedAt:  doc.CreatedAt,
		Status:     doc.Status,
		HPALogs:    logs,
	}

	// resp := models.HPAAlertResponse{
	// 	JobID:      id,
	// 	AppName:    appName,
	// 	Namespace:  namespace,
	// 	Strategy:   "hpa",
	// 	ModifiedAt: mtime,
	// 	CreatedAt:  ctime,
	// 	Status:     status,
	// 	HPALogs:    []string{"test1", "test2", "test3"},
	// }
	return resp
}

// ConvertESToResp .... convert elasticsearch to response
func ConvertESToResp(input models.DocumentResponse) models.ApplicationHealthAnalyzeResponse {

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
	return resp
}

// ConvertESsToResps .... convert elasticsearch results to multiples responses
func ConvertESsToResps(inputs []models.DocumentResponse) []models.ApplicationHealthAnalyzeResponse {

	docs := make([]models.ApplicationHealthAnalyzeResponse, len(inputs))
	fmt.Print(len(inputs))
	for _, element := range inputs {

		docs = append(docs, ConvertESToResp(element))
	}
	return docs
}
