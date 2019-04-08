package search

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	common "foremast.ai/foremast/foremast-service/pkg/common"
	models "foremast.ai/foremast/foremast-service/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
)

const (
	elasticIndexName = "documents"
	elasticTypeName  = "document"
)

// CreateNewDoc .... create new request
func CreateNewDoc(context *gin.Context, elasticClient *elastic.Client, doc models.DocumentRequest) (string, int32) {

	bulk := elasticClient.
		Bulk().
		Index(elasticIndexName).
		Type(elasticTypeName)
	//first search if id already existing.
	id := common.UUIDGen(ConvertDocumentRequestToString(doc))
	log.Println("Generate UUID based on request ", id)
	searchDoc, err, reason := ByID(context, elasticClient, id)

	if err != 0 {
		log.Println("Ignore me, means request is not exist and it is ok to create new request ", searchDoc.ID, "  reason is ", reason)
	}
	if err == -1 {
		var docNew models.Document
		if doc.Strategy == "hpa" {
			docNew = models.Document{
				ID:                    id,
				AppName:               doc.AppName,
				CreatedAt:             time.Now().UTC(),
				StartTime:             common.StrToTime(doc.StartTime),
				EndTime:               common.StrToTime(doc.StartTime),
				ModifiedAt:            time.Now().UTC(),
				CurrentConfig:         doc.CurrentConfig,
				BaselineConfig:        doc.BaselineConfig,
				HistoricalConfig:      doc.HistoricalConfig,
				CurrentMetricStore:    doc.CurrentMetricStore,
				BaselineMetricStore:   doc.BaselineMetricStore,
				HistoricalMetricStore: doc.HistoricalMetricStore,
				Status:                "initial",
				StatusCode:            doc.StatusCode,
				Strategy:              doc.Strategy,
				HPAMetrics:            doc.HPAMetrics,
				Policy:                doc.Policy,
				Namespace:             doc.Namespace,
			}
		} else {
			docNew = models.Document{
				ID:                    id,
				AppName:               doc.AppName,
				CreatedAt:             time.Now().UTC(),
				StartTime:             common.StrToTime(doc.StartTime),
				EndTime:               common.StrToTime(doc.EndTime),
				ModifiedAt:            time.Now().UTC(),
				CurrentConfig:         doc.CurrentConfig,
				BaselineConfig:        doc.BaselineConfig,
				HistoricalConfig:      doc.HistoricalConfig,
				CurrentMetricStore:    doc.CurrentMetricStore,
				BaselineMetricStore:   doc.BaselineMetricStore,
				HistoricalMetricStore: doc.HistoricalMetricStore,
				Status:                "initial",
				StatusCode:            doc.StatusCode,
				Strategy:              doc.Strategy,
			}
		}
		log.Printf("DOC: %#v\n", docNew)
		bulk.Add(elastic.NewBulkIndexRequest().Id(docNew.ID).Doc(docNew))
		if _, err := bulk.Do(context.Request.Context()); err != nil {
			log.Println(err)
			common.ErrorResponse(context, http.StatusInternalServerError, "Failed to create job "+id)
			return id, 0
		}
	}
	return id, 0
}

// ByID .... search elastic search via id or jobid
func ByID(context *gin.Context, elasticClient *elastic.Client, myid string) (models.DocumentResponse, int32, string) {
	skip := 0
	take := 10
	esQuery := elastic.NewMatchQuery("id", myid)
	bQuery := elastic.NewBoolQuery().Must(esQuery)
	result, err := elasticClient.Search().
		Index(elasticIndexName).
		Query(bQuery).
		From(skip).Size(take).
		Do(context)
	if err != nil {
		log.Println(err)
		//common.ErrorResponse(context, http.StatusInternalServerError, "Something went wrong")
		var empty models.DocumentResponse
		return empty, -1, ""
	}
	// Transform search results before returning them
	if len(result.Hits.Hits) == 0 {
		var empty models.DocumentResponse
		return empty, -1, ""
	}
	docs := make([]models.DocumentResponse, 0)
	for _, hit := range result.Hits.Hits {
		var doc models.DocumentResponse
		json.Unmarshal(*hit.Source, &doc)
		docs = append(docs, doc)
	}
	return docs[0], 0, ""

}

// ByQuery .... search by elasticsearch query
func ByQuery(context *gin.Context, elasticClient *elastic.Client, query string) {
	skip := 0
	take := 10
	esQuery := elastic.NewMultiMatchQuery(query, "appName", "content").
		Fuzziness("2").
		MinimumShouldMatch("2")
	result, err := elasticClient.Search().
		Index(elasticIndexName).
		Query(esQuery).
		From(skip).Size(take).
		Do(context)

	if err != nil {
		log.Println(err)
		common.ErrorResponse(context, http.StatusInternalServerError, "Something went wrong")
		return
	}
	// Transform search results before returning them
	docs := make([]models.DocumentResponse, 0)
	for _, hit := range result.Hits.Hits {
		var doc models.DocumentResponse
		json.Unmarshal(*hit.Source, &doc)
		docs = append(docs, doc)
	}
	context.JSON(http.StatusOK, docs)
}

// ByStatus .... This will be used by backend python model, search by open status
func ByStatus(context *gin.Context, elasticClient *elastic.Client, myStatusCode string) {
	skip := 0
	take := 10
	esQuery := elastic.NewMultiMatchQuery(myStatusCode, "statuscode").
		Fuzziness("2").
		MinimumShouldMatch("2")
	result, err := elasticClient.Search().
		Index(elasticIndexName).
		Query(esQuery).
		From(skip).Size(take).
		Do(context)
	if err != nil {
		log.Println(err)
		common.ErrorResponse(context, http.StatusInternalServerError, "Something went wrong")
		return
	}
	// Transform search results before returning them
	docs := make([]models.DocumentResponse, 0)
	for _, hit := range result.Hits.Hits {
		var doc models.DocumentResponse
		json.Unmarshal(*hit.Source, &doc)
		docs = append(docs, doc)
	}
	context.JSON(http.StatusOK, docs)
}

// ConvertDocumentRequestToString ... convert the request to string
func ConvertDocumentRequestToString(doc models.DocumentRequest) string {
	var buffer bytes.Buffer
	buffer.WriteString(doc.AppName)
	buffer.WriteString(doc.StartTime)
	buffer.WriteString(doc.EndTime)
	buffer.WriteString(string(doc.CurrentConfig))
	buffer.WriteString(string(doc.BaselineConfig))
	buffer.WriteString(string(doc.HistoricalConfig))
	buffer.WriteString(string(doc.CurrentMetricStore))
	buffer.WriteString(string(doc.BaselineMetricStore))
	buffer.WriteString(string(doc.HistoricalMetricStore))
	buffer.WriteString(doc.Strategy)
	log.Print("create document request :", buffer.String())
	return buffer.String()
}

/*
func main() {
var err error
// Create Elastic client and wait for Elasticsearch to be ready
for {
	elasticClient, err = elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
	)
	if err != nil {
		log.Println(err)
		// Retry every 3 seconds
		time.Sleep(3 * time.Second)
	} else {
		break
	}
}
	doc := DocumentRequest{
		AppName:   "test1",
		StartTime: "2018-09-12 00:00:00",
		EndTime:   "2018-09-12 00:00:00",
		Content:   "\"k1\":\"v1\",\"k2\":\"v2\"",
	}
*/
//fmt.Println(CreateNewDoc(doc))

//fmt.Println(ByID("*"))

// Start HTTP server
/*
r := gin.Default()
r.POST("/documents", createCheckRequest)
r.GET("/search", searchRequest)
//r.GET("/searchstatus", ByStatus)
if err = r.Run(":8099"); err != nil {
	log.Fatal(err)
}
*
}
func createCheckRequest(context *gin.Context) {
// Parse request
var docs []DocumentRequest
if err := context.BindJSON(&docs); err != nil {
	fmt.Println("enter error block")
	errorResponse(context, http.StatusBadRequest, "Malformed request body")
	return
}
// Insert documents in bulk
bulk := elasticClient.
	Bulk().
	Index(elasticIndexName).
	Type(elasticTypeName)
id := shortid.MustGenerate()
//ids := make([]DocumentResponse, 0)
for _, d := range docs {
	fmt.Println(d.AppName + " " + d.Content + "  " + id)
	doc := Document{
		ID:         id,
		AppName:    d.AppName,
		CreatedAt:  time.Now().UTC(),
		StartTime:  StrToTime(d.StartTime),
		EndTime:    StrToTime(d.EndTime),
		ModifiedAt: time.Now().UTC(),
		Content:    d.Content,
		Status:     "created",
		Stage:      "initial",
	}
	bulk.Add(elastic.NewBulkIndexRequest().Id(doc.ID).Doc(doc))
}
if _, err := bulk.Do(context.Request.Context()); err != nil {
	log.Println(err)
	errorResponse(context, http.StatusInternalServerError, "Failed to create documents")
	return
}
context.Status(http.StatusOK)
}
func searchRequest(context *gin.Context) {
// Parse request
query := context.Query("query")
fmt.Println(query)
col := "query"
if query == "" {
	query = context.Query("id")
	col = "id"
	if query == "" {
		errorResponse(context, http.StatusBadRequest, "Query not specified")
		return
	}
}
fmt.Println(col)
fmt.Println(query)
skip := 0
take := 10
if i, err := strconv.Atoi(context.Query("skip")); err == nil {
	skip = i
}
if i, err := strconv.Atoi(context.Query("take")); err == nil {
	take = i
}
// Perform search
if  col == "query" {
	esQuery := elastic.NewMultiMatchQuery(query, "appName", "content").
		Fuzziness("2").
		MinimumShouldMatch("2")
	result, err := elasticClient.Search().
		Index(elasticIndexName).
		Query(esQuery).
		From(skip).Size(take).
		Do(context.Request.Context())
	if err != nil {
		log.Println(err)
		errorResponse(context, http.StatusInternalServerError, "Something went wrong")
		return
	}
	res := SearchResponse{
		Time: fmt.Sprintf("%d", result.TookInMillis),
		Hits: fmt.Sprintf("%d", result.Hits.TotalHits),
	}
	// Transform search results before returning them
	docs := make([]DocumentResponse, 0)
	for _, hit := range result.Hits.Hits {
		var doc DocumentResponse
		json.Unmarshal(*hit.Source, &doc)
		docs = append(docs, doc)
	}
	res.Documents = docs
	context.JSON(http.StatusOK, res)
	return;
}
	esQuery := elastic.NewMultiMatchQuery(query, "id").
		Fuzziness("2").
		MinimumShouldMatch("2")
	result, err := elasticClient.Search().
		Index(elasticIndexName).
		Query(esQuery).
		From(skip).Size(take).
		Do(context.Request.Context())
if err != nil {
	log.Println(err)
	errorResponse(c, http.StatusInternalServerError, "Something went wrong")
	return
}
res := SearchResponse{
	Time: fmt.Sprintf("%d", result.TookInMillis),
	Hits: fmt.Sprintf("%d", result.Hits.TotalHits),
}
// Transform search results before returning them
docs := make([]DocumentResponse, 0)
for _, hit := range result.Hits.Hits {
	var doc DocumentResponse
	json.Unmarshal(*hit.Source, &doc)
	docs = append(docs, doc)
}
res.Documents = docs
context.JSON(http.StatusOK, res)
}
*/
