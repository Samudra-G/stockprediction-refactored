package pkg

import (
	"math"
	"testing"
)

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestMovingAverage(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	window := 3
	expected := []float64{2, 3, 4}

	got := MovingAverage(data, window)
	if len(got) != len(expected) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(expected))
	}

	for i := range got {
		if !almostEqual(got[i], expected[i], 1e-6) {
			t.Errorf("at index %d: got %v, want %v", i, got[i], expected[i])
		}
	}
}

func TestRSI(t *testing.T) {
	data := []float64{44, 44.15, 43.9, 44.35, 44.8, 45, 45.2, 45.05, 44.9, 45.1, 45.3, 45.5, 45.8, 46, 46.2, 46.4}
	period := 5

	got := RSI(data, period)
	if len(got) == 0 {
		t.Errorf("RSI() returned nil or empty, got %v", got)
	}
}

func TestVolatility(t *testing.T) {
	data := []float64{100, 102, 101, 105, 107}

	got := Volatility(data)
	if got <= 0 {
		t.Errorf("Volatility() = %v, want > 0", got)
	}
}

func TestEMA(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	period := 3

	got := EMA(data, period)
	if len(got) != len(data) {
		t.Errorf("EMA() length = %v, want %v", len(got), len(data))
	}
}

func TestMACD(t *testing.T) {
	data := []float64{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
	}

	macdLine, signalLine, histogram := MACD(data)

	if len(macdLine) != len(data) {
		t.Errorf("MACD line length = %v, want %v", len(macdLine), len(data))
	}
	if len(signalLine) != len(data) {
		t.Errorf("Signal line length = %v, want %v", len(signalLine), len(data))
	}
	if len(histogram) != len(data) {
		t.Errorf("Histogram length = %v, want %v", len(histogram), len(data))
	}
}