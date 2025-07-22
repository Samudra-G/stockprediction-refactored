package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Setup test environment
func setupTest() (*Handler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler()
	router := gin.New()
	
	router.GET("/health", handler.Health)
	router.POST("/metric", handler.Metric)
	router.GET("/poll", handler.Poll)
	
	return handler, router
}

// Create multipart form with CSV file
func createMultipartForm(csvData, ticker string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add ticker field
	if ticker != "" {
		err := writer.WriteField("ticker", ticker)
		if err != nil {
			return nil, "", err
		}
	}

	// Add file field
	part, err := writer.CreateFormFile("file", "test.csv")
	if err != nil {
		return nil, "", err
	}

	_, err = part.Write([]byte(csvData))
	if err != nil {
		return nil, "", err
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	return body, contentType, nil
}

func TestHandler_Health(t *testing.T) {
	_, router := setupTest()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Go Backend running...", response["status"])
}

func TestHandler_Metric_Success(t *testing.T) {
	_, router := setupTest()

	// Create test CSV data with sufficient data points for calculations
	csvData := `Date,Open,High,Low,Close,Volume
2023-01-01,100.0,105.0,95.0,102.0,1000000
2023-01-02,102.0,107.0,100.0,105.0,1100000
2023-01-03,105.0,110.0,103.0,108.0,1200000
2023-01-04,108.0,112.0,106.0,110.0,1300000
2023-01-05,110.0,115.0,108.0,113.0,1400000`

	// Add more data points to ensure we have enough for MA100, MA200
	for i := 6; i <= 250; i++ {
		csvData += fmt.Sprintf("\n2023-01-%02d,%.1f,%.1f,%.1f,%.1f,%d",
			i%30+1, 100.0+float64(i), 105.0+float64(i), 95.0+float64(i), 100.0+float64(i), 1000000+i*1000)
	}

	body, contentType, err := createMultipartForm(csvData, "AAPL")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/metric", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check that all expected fields are present
	assert.Contains(t, response, "ma100")
	assert.Contains(t, response, "ma200")
	assert.Contains(t, response, "rsi")
	assert.Contains(t, response, "volatility")
	assert.Contains(t, response, "macd")
	assert.Contains(t, response, "signal")
	assert.Contains(t, response, "histogram")

	// Verify the response types
	assert.IsType(t, []interface{}{}, response["ma100"])
	assert.IsType(t, []interface{}{}, response["ma200"])
	assert.IsType(t, []interface{}{}, response["rsi"])
	assert.IsType(t, float64(0), response["volatility"])
}

func TestHandler_Metric_MissingFile(t *testing.T) {
	_, router := setupTest()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("ticker", "AAPL")
	writer.Close()

	req, _ := http.NewRequest("POST", "/metric", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "CSV file is required", response["error"])
}

func TestHandler_Metric_MissingTicker(t *testing.T) {
	_, router := setupTest()

	csvData := `Date,Open,High,Low,Close,Volume
2023-01-01,100.0,105.0,95.0,102.0,1000000`

	body, contentType, err := createMultipartForm(csvData, "")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/metric", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Ticker is required", response["error"])
}

func TestHandler_Metric_InvalidCSV(t *testing.T) {
	_, router := setupTest()

	// CSV without Close column
	csvData := `Date,Open,High,Low,Volume
2023-01-01,100.0,105.0,95.0,1000000`

	body, contentType, err := createMultipartForm(csvData, "AAPL")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/metric", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Close")
}

func TestHandler_Metric_InsufficientData(t *testing.T) {
	_, router := setupTest()

	// CSV with only one data point
	csvData := `Date,Open,High,Low,Close,Volume
2023-01-01,100.0,105.0,95.0,102.0,1000000`

	body, contentType, err := createMultipartForm(csvData, "AAPL")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/metric", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "not enough close prices")
}

func TestHandler_Poll_MissingTicker(t *testing.T) {
	_, router := setupTest()

	req, _ := http.NewRequest("GET", "/poll", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ticker parameter required", response["error"])
}

/*func TestHandler_Poll_NoExistingChannel(t *testing.T) {
	_, router := setupTest()

	req, _ := http.NewRequest("GET", "/poll?ticker=AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "idle", response["status"])
	assert.Nil(t, response["predictions"])
}*/

func TestHandler_Poll_PendingPrediction(t *testing.T) {
	handler, router := setupTest()

	// Create a prediction channel and store it
	predictionCh := make(chan PredictionResponse, 1)
	handler.storePredictionChannel("AAPL", predictionCh)

	req, _ := http.NewRequest("GET", "/poll?ticker=AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "pending", response["status"])
	assert.Nil(t, response["predictions"])

	// Clean up
	close(predictionCh)
}

func TestHandler_Poll_CompletedPrediction(t *testing.T) {
	handler, router := setupTest()

	// Create a prediction channel with a result
	predictionCh := make(chan PredictionResponse, 1)
	testData := map[string]interface{}{"prediction": 123.45}
	predictionCh <- PredictionResponse{
		Status: "success",
		Data:   testData,
		Error:  nil,
	}
	handler.storePredictionChannel("AAPL", predictionCh)

	req, _ := http.NewRequest("GET", "/poll?ticker=AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["predictions"])

	// Verify channel is removed after consumption
	_, exists := handler.getPredictionChannel("AAPL")
	assert.False(t, exists)

	close(predictionCh)
}

func TestHandler_Poll_FailedPrediction(t *testing.T) {
	handler, router := setupTest()

	// Create a prediction channel with an error result
	predictionCh := make(chan PredictionResponse, 1)
	predictionCh <- PredictionResponse{
		Status: "failed",
		Data:   nil,
		Error:  fmt.Errorf("prediction failed"),
	}
	handler.storePredictionChannel("AAPL", predictionCh)

	req, _ := http.NewRequest("GET", "/poll?ticker=AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "failed", response["status"])
	assert.Nil(t, response["predictions"])
	assert.Equal(t, "prediction failed", response["error"])

	close(predictionCh)
}

func TestPredictionService_RequestPrediction(t *testing.T) {
	// Start a fake ML backend server
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/predict", r.URL.Path)
		require.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"predictions":{"predictions":[100.1,101.2],"y_test":[99.8,100.5],"dates":["2025-07-20","2025-07-21"]}}`))
	}))
	defer fakeServer.Close()

	// Set env variable so your handler picks this fake backend
	t.Setenv("ML_BACKEND", fakeServer.URL)

	// 🔧 Create the PredictionService
	ps := NewPredictionService(1)
	defer close(ps.requestCh)

	testData := []byte("Close,Other\n100,foo\n101,foo\n102,foo\n")
	testFileName := "test.csv"

	// Send request
	responseCh := ps.RequestPrediction(testData, testFileName)
	assert.NotNil(t, responseCh)

	// Wait for response
	select {
	case response := <-responseCh:
		assert.Equal(t, "success", response.Status)
		assert.Nil(t, response.Error)

		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok, "data should be a map")

		preds, ok := data["predictions"]
		assert.True(t, ok, "predictions key should be present")
		assert.NotNil(t, preds)

	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for prediction response")
	}
}

func TestPredictionService_ChannelFull(t *testing.T) {
	// Create a service with a small buffer and no workers
	ps := &PredictionService{
		requestCh: make(chan PredictionRequest, 1), // Small buffer
		workers:   0, // No workers to process requests
	}

	testData := []byte("test data")
	testFileName := "test.csv"

	// Fill the channel
	ps.RequestPrediction(testData, testFileName)
	
	// This should return immediately with an error
	responseCh := ps.RequestPrediction(testData, testFileName)
	
	select {
	case response := <-responseCh:
		assert.Equal(t, "failed", response.Status)
		assert.Contains(t, response.Error.Error(), "busy")
	case <-time.After(1 * time.Second):
		t.Fatal("Expected immediate response when channel is full")
	}
}

// Integration test that combines metric calculation and polling
func TestHandler_MetricAndPoll_Integration(t *testing.T) {
	// Skip this test if ML_BACKEND is not set (to avoid FastAPI calls in unit tests)
	if os.Getenv("ML_BACKEND") == "" {
		t.Skip("Skipping integration test: ML_BACKEND not set")
	}

	_, router := setupTest()

	// Create test data
	csvData := `Date,Open,High,Low,Close,Volume
2023-01-01,100.0,105.0,95.0,102.0,1000000
2023-01-02,102.0,107.0,100.0,105.0,1100000`

	// Add more data points
	for i := 3; i <= 250; i++ {
		csvData += fmt.Sprintf("\n2023-01-%02d,%.1f,%.1f,%.1f,%.1f,%d",
			i%30+1, 100.0+float64(i), 105.0+float64(i), 95.0+float64(i), 100.0+float64(i), 1000000+i*1000)
	}

	// First, call metric endpoint
	body, contentType, err := createMultipartForm(csvData, "TSLA")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/metric", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Then poll for predictions
	req, _ = http.NewRequest("GET", "/poll?ticker=TSLA", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should be either pending or completed (depending on timing)
	status := response["status"]
	assert.True(t, status == "pending" || status == "failed" || status == "success")
}

// Benchmark tests
func BenchmarkHandler_Metric(b *testing.B) {
	_, router := setupTest()

	csvData := `Date,Open,High,Low,Close,Volume`
	for i := 1; i <= 250; i++ {
		csvData += fmt.Sprintf("\n2023-01-%02d,%.1f,%.1f,%.1f,%.1f,%d",
			i%30+1, 100.0+float64(i), 105.0+float64(i), 95.0+float64(i), 100.0+float64(i), 1000000+i*1000)
	}

	body, contentType, _ := createMultipartForm(csvData, "AAPL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/metric", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}