package scraper

import "math"

func GetBatchSize(total, batches int) int {
	return int(math.Ceil(float64(total) / float64(batches)))
}

func AddInt(x, y int) int {
	return x + y
}

func AddFloat(x, y float64) float64 {
	return x + y
}

func Multiply(x, y float64) float64 {
	return x * y
}

func ToFloat(x int) float64 {
	return float64(x)
}

func GetQuestionDifficultyColour(difficulty string) string {
	switch difficulty {
	case "Easy":
		return "#46c6c2"
	case "Medium":
		return "#ffc01e"
	case "Hard":
		return "#ff375f"
	}
	return ""
}
