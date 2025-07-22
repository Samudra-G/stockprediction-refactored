package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Samudra-G/stockprediction-refactored/pkg"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

// Removed unused global context

type Handler struct {
	predictionService *PredictionService
}

func NewHandler() *Handler {
	return &Handler{
		predictionService: NewPredictionService(3), // 3 workers for FastAPI calls
	}
}

// PredictionRequest represents a prediction request
type PredictionRequest struct {
	FileData   []byte
	FileName   string
	ResponseCh chan PredictionResponse
}

// PredictionResponse represents a prediction response
type PredictionResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  error       `json:"error,omitempty"`
}

// PredictionService handles prediction requests using channels
type PredictionService struct {
	requestCh chan PredictionRequest
	workers   int
}

func NewPredictionService(workers int) *PredictionService {
	ps := &PredictionService{
		requestCh: make(chan PredictionRequest, 10), // buffered channel
		workers:   workers,
	}

	// Start worker pool
	for i := 0; i < workers; i++ {
		go ps.worker()
	}

	return ps
}

func (ps *PredictionService) worker() {
	for req := range ps.requestCh {
		result := ps.callFastAPI(req.FileData, req.FileName)
		req.ResponseCh <- result
		close(req.ResponseCh)
	}
}

func (ps *PredictionService) callFastAPI(fileData []byte, fileName string) PredictionResponse {
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		log.Println("Failed to create form file:", err)
		return PredictionResponse{Status: "failed", Error: err}
	}

	_, err = part.Write(fileData)
	if err != nil {
		log.Println("Failed to write file contents:", err)
		return PredictionResponse{Status: "failed", Error: err}
	}

	writer.Close()

	mlBackendURL := os.Getenv("ML_BACKEND")
	if mlBackendURL == "" {
		log.Println("ML_BACKEND environment variable is not set")
		return PredictionResponse{Status: "failed", Error: err}
	}

	req, err := http.NewRequest("POST", mlBackendURL+"/api/v1/predict", bodyBuf)
	if err != nil {
		log.Println("Failed to create FastAPI request:", err)
		return PredictionResponse{Status: "failed", Error: err}
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("FastAPI request failed:", err)
		return PredictionResponse{Status: "failed", Error: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("FastAPI returned bad status:", resp.Status)
		return PredictionResponse{Status: "failed", Error: err}
	}

	var result map[string]interface{}
	if err := pkg.FromJSON(resp.Body, &result); err != nil {
		log.Println("Failed to decode FastAPI response:", err)
		return PredictionResponse{Status: "failed", Error: err}
	}

	return PredictionResponse{Status: "success", Data: result}
}

// RequestPrediction submits a prediction request and returns a response channel
func (ps *PredictionService) RequestPrediction(fileData []byte, fileName string) <-chan PredictionResponse {
	responseCh := make(chan PredictionResponse, 1)
	
	select {
	case ps.requestCh <- PredictionRequest{
		FileData:   fileData,
		FileName:   fileName,
		ResponseCh: responseCh,
	}:
		return responseCh
	default:
		// Channel is full, reject request
		go func() {
			responseCh <- PredictionResponse{
				Status: "failed",
				Error:  fmt.Errorf("prediction service busy, try again later"),
			}
			close(responseCh)
		}()
		return responseCh
	}
}

// MetricResults holds all calculated metrics
type MetricResults struct {
	MA100      []float64   `json:"ma100"`
	MA200      []float64   `json:"ma200"`
	RSI        []float64   `json:"rsi"`
	Volatility float64     `json:"volatility"`
	MACD       []float64   `json:"macd"`
	Signal     []float64   `json:"signal"`
	Histogram  []float64   `json:"histogram"`
	PredictionCh <-chan PredictionResponse `json:"-"`
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "Go Backend running..."})
}

func (h *Handler) Metric(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("Failed to get form file:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file is required"})
		return
	}

	ticker := c.PostForm("ticker")
	if ticker == "" {
		log.Println("Ticker not provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticker is required"})
		return
	}

	// Parse CSV and calculate metrics concurrently
	results, err := h.processMetrics(file)
	if err != nil {
		log.Println("Failed to process metrics:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store prediction channel for polling (you might want to use Redis here later)
	// For now, we'll still use a simple in-memory store but with better structure
	h.storePredictionChannel(ticker, results.PredictionCh)

	c.JSON(http.StatusOK, gin.H{
		"ma100":      results.MA100,
		"ma200":      results.MA200,
		"rsi":        results.RSI,
		"volatility": results.Volatility,
		"macd":       results.MACD,
		"signal":     results.Signal,
		"histogram":  results.Histogram,
	})
}

func (h *Handler) processMetrics(file *multipart.FileHeader) (*MetricResults, error) {
	// Read and parse CSV
	closes, fileData, err := h.parseCSV(file)
	if err != nil {
		return nil, err
	}

	if len(closes) < 2 {
		return nil, fmt.Errorf("not enough close prices")
	}

	// Create errgroup for concurrent metric calculations
	g := &errgroup.Group{}
	
	var (
		ma100     []float64
		ma200     []float64
		rsi       []float64
		vol       float64
		macdLine  []float64
		signalLine []float64
		histogram []float64
	)

	// Calculate metrics concurrently
	g.Go(func() error {
		ma100 = pkg.MovingAverage(closes, 100)
		return nil
	})

	g.Go(func() error {
		ma200 = pkg.MovingAverage(closes, 200)
		return nil
	})

	g.Go(func() error {
		rsi = pkg.RSI(closes, 14)
		return nil
	})

	g.Go(func() error {
		vol = pkg.Volatility(closes)
		return nil
	})

	g.Go(func() error {
		macdLine, signalLine, histogram = pkg.MACD(closes)
		return nil
	})

	// Wait for all metric calculations to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Start prediction request (non-blocking)
	predictionCh := h.predictionService.RequestPrediction(fileData, file.Filename)

	return &MetricResults{
		MA100:        ma100,
		MA200:        ma200,
		RSI:          rsi,
		Volatility:   vol,
		MACD:         macdLine,
		Signal:       signalLine,
		Histogram:    histogram,
		PredictionCh: predictionCh,
	}, nil
}

func (h *Handler) parseCSV(file *multipart.FileHeader) ([]float64, []byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer f.Close()

	// Read file data for FastAPI (we need this as bytes)
	fileData, err := io.ReadAll(f)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file data: %w", err)
	}

	// Reset reader for CSV parsing
	f.Seek(0, 0)
	reader := csv.NewReader(f)

	// Reading header
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Find "Close" column
	closeIdx := -1
	for i, col := range headers {
		if col == "Close" {
			closeIdx = i
			break
		}
	}
	if closeIdx == -1 {
		return nil, nil, fmt.Errorf(`"Close" column not found in CSV headers`)
	}

	var closes []float64
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error reading CSV row: %w", err)
		}

		val, err := strconv.ParseFloat(record[closeIdx], 64)
		if err != nil {
			log.Println("Failed to parse float value:", record[closeIdx], err)
			continue
		}

		closes = append(closes, val)
	}

	return closes, fileData, nil
}

// Simple in-memory store for prediction channels (replace with Redis later)
var (
	predictionChannels = make(map[string]<-chan PredictionResponse)
	channelMutex       = &sync.RWMutex{}
)

func (h *Handler) storePredictionChannel(ticker string, ch <-chan PredictionResponse) {
	channelMutex.Lock()
	defer channelMutex.Unlock()
	predictionChannels[ticker] = ch
}

func (h *Handler) getPredictionChannel(ticker string) (<-chan PredictionResponse, bool) {
	channelMutex.RLock()
	defer channelMutex.RUnlock()
	ch, exists := predictionChannels[ticker]
	return ch, exists
}

func (h *Handler) Poll(c *gin.Context) {
	ticker := c.Query("ticker")
	if ticker == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticker parameter required"})
		return
	}

	ch, exists := h.getPredictionChannel(ticker)
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"status":      "idle",
			"predictions": nil,
		})
		return
	}

	// Non-blocking check for prediction result
	select {
	case result := <-ch:
		// Remove from store after consuming
		channelMutex.Lock()
		delete(predictionChannels, ticker)
		channelMutex.Unlock()

		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":      "failed",
				"predictions": nil,
				"error":       result.Error.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":      result.Status,
				"predictions": result.Data,
			})
		}
	default:
		// Still pending
		c.JSON(http.StatusOK, gin.H{
			"status":      "pending",
			"predictions": nil,
		})
	}
}