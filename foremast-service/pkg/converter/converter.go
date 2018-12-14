package converter

import (
	models "foremast.ai/foremast/foremast-service/pkg/models"
	"fmt"
	"strconv"
)

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

func ConvertESToNewResp(uuid string, statusCode int32, status, reason string) models.ApplicationHealthAnalyzeResponseNew {
	if statusCode == 0 {
		statusCode = 200
	}

	resp := models.ApplicationHealthAnalyzeResponseNew{
		JobId:      uuid,
		StatusCode: statusCode,
		Status:     status,
		Reason:     reason,
	}
	return resp
}

func ConvertESToResp(input models.DocumentResponse) models.ApplicationHealthAnalyzeResponse {

	code, err := strconv.Atoi(input.StatusCode)
	if err != nil {
		code = 200
	}
	resp := models.ApplicationHealthAnalyzeResponse{
		JobId:      input.ID,
		StatusCode: int32(code),
		Status:     ConvertStatusToExternal(input.Status),
		Reason:     input.Reason,
	}
	return resp
}

func ConvertESsToResps(inputs []models.DocumentResponse) []models.ApplicationHealthAnalyzeResponse {

	docs := make([]models.ApplicationHealthAnalyzeResponse, len(inputs))
	fmt.Print(len(inputs))
	for _, element := range inputs {

		docs = append(docs, ConvertESToResp(element))
	}
	return docs
}
