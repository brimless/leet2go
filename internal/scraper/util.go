package scraper

import "math"

func GetBatchSize(total, batches int) int {
	return int(math.Ceil(float64(total) / float64(batches)))
}
