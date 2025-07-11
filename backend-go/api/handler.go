package api

import (
	"bytes"
	"context"
	"encoding/csv"
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
)

var ctx = NewContext()

func NewContext() (ctx context.Context) {
	return context.Background()
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type PredictionState struct {
	Status      string      // "pending", "success", "failed"
	Predictions interface{}
}

var (
	predictionState = PredictionState{
		Status:      "idle",
		Predictions: nil,
	}
	stateLock = &sync.Mutex{}
)

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

	f, err := file.Open()
	if err != nil {
		log.Println("Failed to open uploaded file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open CSV"})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Reading header
	headers, err := reader.Read()
	if err != nil {
		log.Println("Failed to read CSV header:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read CSV header"})
		return
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
		log.Println(`"Close" column not found in CSV headers`, headers)
		c.JSON(http.StatusBadRequest, gin.H{"error": `"Close" column not found`})
		return
	}

	var closes []float64
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error reading CSV row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse CSV rows"})
			return
		}

		val, err := strconv.ParseFloat(record[closeIdx], 64)
		if err != nil {
			log.Println("Failed to parse float value:", record[closeIdx], err)
			continue
		}

		closes = append(closes, val)
	}

	if len(closes) < 2 {
		log.Println("Not enough close prices:", closes)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough Close prices"})
		return
	}

	ma100 := pkg.MovingAverage(closes, 100)
	ma200 := pkg.MovingAverage(closes, 200)
	rsi := pkg.RSI(closes, 14)
	vol := pkg.Volatility(closes)
	macdLine, signalLine, histogram := pkg.MACD(closes)

	// Re-open file for FastAPI
	f2, err := file.Open()
	if err != nil {
		log.Println("Failed to reopen file for FastAPI:", err)
		stateLock.Lock()
		predictionState.Status = "failed"
		stateLock.Unlock()
	} else {
		defer f2.Close()

		stateLock.Lock()
		predictionState.Status = "pending"
		predictionState.Predictions = nil
		stateLock.Unlock()

		go func() {
			log.Println("Calling FastAPI with CSV file for prediction...")

			bodyBuf := &bytes.Buffer{}
			writer := multipart.NewWriter(bodyBuf)

			part, err := writer.CreateFormFile("file", file.Filename)
			if err != nil {
				log.Println("Failed to create form file:", err)
				stateLock.Lock()
				predictionState.Status = "failed"
				stateLock.Unlock()
				return
			}

			_, err = io.Copy(part, f2)
			if err != nil {
				log.Println("Failed to copy file contents:", err)
				stateLock.Lock()
				predictionState.Status = "failed"
				stateLock.Unlock()
				return
			}

			writer.Close()

			mlBackendURL := os.Getenv("ML_BACKEND")
			if mlBackendURL == "" {
				log.Println("ML_BACKEND environment variable is not set")
				stateLock.Lock()
				predictionState.Status = "failed"
				stateLock.Unlock()
				return
			}

			req, err := http.NewRequest("POST", mlBackendURL+"/api/v1/predict", bodyBuf)
			if err != nil {
				log.Println("Failed to create FastAPI request:", err)
				stateLock.Lock()
				predictionState.Status = "failed"
				stateLock.Unlock()
				return
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			client := &http.Client{
				Timeout: 60 * time.Second,
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Println("FastAPI request failed:", err)
				stateLock.Lock()
				predictionState.Status = "failed"
				stateLock.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Println("FastAPI returned bad status:", resp.Status)
				stateLock.Lock()
				predictionState.Status = "failed"
				stateLock.Unlock()
				return
			}

			var result map[string]interface{}
			if err := pkg.FromJSON(resp.Body, &result); err != nil {
				log.Println("Failed to decode FastAPI response:", err)
				stateLock.Lock()
				predictionState.Status = "failed"
				stateLock.Unlock()
				return
			}

			stateLock.Lock()
			predictionState.Status = "success"
			predictionState.Predictions = result
			stateLock.Unlock()
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"ma100":      ma100,
		"ma200":      ma200,
		"rsi":        rsi,
		"volatility": vol,
		"macd":       macdLine,
		"signal":     signalLine,
		"histogram":  histogram,
	})
}

func (h *Handler) Poll(c *gin.Context) {
	stateLock.Lock()
	defer stateLock.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"status":      predictionState.Status,
		"predictions": predictionState.Predictions,
	})

	if predictionState.Status == "success" || predictionState.Status == "failed" {
		// Auto-reset for next run
		predictionState.Status = "idle"
		predictionState.Predictions = nil
	}
}
