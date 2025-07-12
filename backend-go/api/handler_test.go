package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMetricSuccess(t *testing.T) {
	defer resetState()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.POST("/metric", h.Metric)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", "test.csv")
	io := strings.NewReader("Close,Other\n100,foo\n101,foo\n102,foo\n")
	_, _ = io.WriteTo(fw)
	_ = w.WriteField("ticker", "FAKE")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/metric", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()

	os.Setenv("ML_BACKEND", "")

	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	body := resp.Body.String()
	require.Contains(t, body, "ma100")
	require.Contains(t, body, "rsi")
	require.Contains(t, body, "macd")
}

func TestMetricMissingFile(t *testing.T) {
	defer resetState()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.POST("/metric", h.Metric)

	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	_ = w.WriteField("ticker", "FAKE")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/metric", b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "CSV file is required")
}

func TestMetricMissingTicker(t *testing.T) {
	defer resetState()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.POST("/metric", h.Metric)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", "test.csv")
	io := strings.NewReader("Close,Other\n100,foo\n101,foo\n")
	_, _ = io.WriteTo(fw)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/metric", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "Ticker is required")
}

func TestMetricInvalidHeader(t *testing.T) {
	defer resetState()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.POST("/metric", h.Metric)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", "test.csv")
	io := strings.NewReader("Price,Other\n100,foo\n101,foo\n")
	_, _ = io.WriteTo(fw)
	_ = w.WriteField("ticker", "FAKE")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/metric", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "Close column not found")
}

func TestMetricNotEnoughCloses(t *testing.T) {
	defer resetState()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.POST("/metric", h.Metric)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", "test.csv")
	io := strings.NewReader("Close,Other\n100,foo\n")
	_, _ = io.WriteTo(fw)
	_ = w.WriteField("ticker", "FAKE")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/metric", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "Not enough Close prices")
}

func TestPollIdle(t *testing.T) {
	defer resetState()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.GET("/poll", h.Poll)

	req := httptest.NewRequest(http.MethodGet, "/poll", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Contains(t, resp.Body.String(), "idle")
}

func TestPollAutoReset(t *testing.T) {
	defer resetState()

	stateLock.Lock()
	predictionState.Status = "success"
	predictionState.Predictions = map[string]interface{}{"foo": "bar"}
	stateLock.Unlock()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.GET("/poll", h.Poll)

	req := httptest.NewRequest(http.MethodGet, "/poll", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Contains(t, resp.Body.String(), "success")

	// Second call should be idle
	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req)
	require.Contains(t, resp2.Body.String(), "idle")
}

func TestHealth(t *testing.T) {
	defer resetState()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.GET("/health", h.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Contains(t, resp.Body.String(), "Go Backend running")
}

func TestMetricFastAPISuccess(t *testing.T) {
	defer resetState()

	// Start fake FastAPI server that returns 200 OK with JSON
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/predict", r.URL.Path)
		require.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"mocked_prediction"}`))
	}))
	defer fakeServer.Close()

	os.Setenv("ML_BACKEND", fakeServer.URL)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.POST("/metric", h.Metric)
	r.GET("/poll", h.Poll)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", "test.csv")
	io := strings.NewReader("Close,Other\n100,foo\n101,foo\n102,foo\n")
	_, _ = io.WriteTo(fw)
	_ = w.WriteField("ticker", "FAKE")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/metric", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	// Poll until success
	var pollResp *httptest.ResponseRecorder
	success := false
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)

		reqPoll := httptest.NewRequest(http.MethodGet, "/poll", nil)
		pollResp = httptest.NewRecorder()
		r.ServeHTTP(pollResp, reqPoll)

		if strings.Contains(pollResp.Body.String(), "success") {
			success = true
			break
		}
	}

	require.True(t, success, "expected poll to eventually return success")
	require.Contains(t, pollResp.Body.String(), "mocked_prediction")
}

func TestMetricFastAPIFailure(t *testing.T) {
	defer resetState()

	// Start fake FastAPI server that returns 500
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer fakeServer.Close()

	os.Setenv("ML_BACKEND", fakeServer.URL)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := NewHandler()
	r.POST("/metric", h.Metric)
	r.GET("/poll", h.Poll)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", "test.csv")
	io := strings.NewReader("Close,Other\n100,foo\n101,foo\n102,foo\n")
	_, _ = io.WriteTo(fw)
	_ = w.WriteField("ticker", "FAKE")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/metric", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	// Poll until failed
	var pollResp *httptest.ResponseRecorder
	failed := false
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)

		reqPoll := httptest.NewRequest(http.MethodGet, "/poll", nil)
		pollResp = httptest.NewRecorder()
		r.ServeHTTP(pollResp, reqPoll)

		if strings.Contains(pollResp.Body.String(), "failed") {
			failed = true
			break
		}
	}

	require.True(t, failed, "expected poll to eventually return failed")
}

// Helper
func resetState() {
	stateLock.Lock()
	predictionState.Status = "idle"
	predictionState.Predictions = nil
	stateLock.Unlock()
}
