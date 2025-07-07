package pkg

import (
	"math"

	"github.com/Samudra-G/stockprediction-refactored/utils"
)

func MovingAverage(data []float64, window int) []float64 {
	if len(data) < window {
		return nil
	}

	ma := make([]float64, len(data)-window+1)
	for i:=0; i <= len(data)-window; i++ {
		sum := 0.0
		for j := i; j < i+window; j++ {
			sum += data[j]
		}
		ma[i] = sum / float64(window)
	}
	return ma
}

func RSI(prices []float64, period int) []float64 {
	if len(prices) <= period {
		return nil
	}

	gains := make([]float64, len(prices)-1)
	losses := make([]float64, len(prices)-1)

	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains[i-1] = change
		} else {
			losses[i-1] = change
		}
	}
	
	avgGain := utils.Average(gains[:period])
	avgLoss := utils.Average(losses[:period])

	rsi := make([]float64, len(prices)-period)
	if avgLoss == 0 {
		rsi[0] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[0] = 100 - (100 / (1 + rs))
	}

	for i := period; i < len(gains); i++ {
		avgGain = ((avgGain * float64(period-1)) + gains[i]) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + losses[i]) / float64(period)

		if avgLoss == 0 {
			rsi[i-period+1] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i-period+1] = 100 - (100 / (1 + rs))
		}
	}

	return rsi
}

func Volatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	// daily returns
	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
	}

	mean := utils.Average(returns)
	var sumSq float64
	for _, r := range returns {
		sumSq += math.Pow(r-mean, 2)
	}

	variance := sumSq / float64(len(returns)-1)
	return math.Sqrt(variance)
}

func EMA(data []float64, period int) []float64 {
    ema := make([]float64, len(data))
    multiplier := 2.0 / float64(period+1)

    sum := 0.0
    for i := 0; i < period && i < len(data); i++ {
        sum += data[i]
    }
    if len(data) < period {
        return ema // return nil if not enough data
    }

    ema[period-1] = sum / float64(period)

    for i := period; i < len(data); i++ {
        ema[i] = (data[i]-ema[i-1])*multiplier + ema[i-1]
    }
    return ema
}

// MACD returns the MACD line, Signal line, and Histogram
func MACD(data []float64) ([]float64, []float64, []float64) {
    ema12 := EMA(data, 12)
    ema26 := EMA(data, 26)

    macdLine := make([]float64, len(data))
    for i := 0; i < len(data); i++ {
        macdLine[i] = ema12[i] - ema26[i]
    }

    signalLine := EMA(macdLine, 9)

    histogram := make([]float64, len(data))
    for i := 0; i < len(data); i++ {
        histogram[i] = macdLine[i] - signalLine[i]
    }

    return macdLine, signalLine, histogram
}