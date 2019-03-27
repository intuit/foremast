package analyst

import (
	"testing"
)

func TestPayload(*testing.T) {

	//var client, err = NewClient(nil, "http://localhost:8099/v1/dhealthcheck/")
	//if err != nil {
	//	glog.Error(err)
	//	return
	//}
	//
	//jobId, err := client.StartAnalyzing("default", "k8s-metrics-demo", nil, "", d.Metrics{})
	//if err != nil {
	//	glog.Error(err)
	//	return
	//}
	//
	//glog.Info(jobId)
	//
	//getStatusResponse, err := client.GetStatus(jobId)
	//if err != nil {
	//	glog.Error(err)
	//	return
	//}
	//
	//glog.Info(getStatusResponse.Status)
}

// func TestTimeFormat(*testing.T) {
//
// 	var t = time.Now()
// 	var start = t.Format(time.RFC3339)
// 	var waitUntil = t.Add(30 * time.Minute).Format(time.RFC3339)
// 	glog.Info(start)
// 	glog.Info(waitUntil)
// }
