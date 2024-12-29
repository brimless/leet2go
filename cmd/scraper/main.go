package main

import (
	"context"
	"log"
	"sync"

	"github.com/brimless/leet2go/internal/scraper"
)

func main() {
	ctx := context.Background()

	// create an instance of the http client we're going to use
	// can use or not use proxies
	client := scraper.NewLeetCodeClient(scraper.USE_PROXY)

	// get the total amount of questions on leetcode (to run and batch requests in parallel)
	totalQuestionsCount, err := client.FetchQuestionsCount()
	if err != nil {
		log.Fatalf("failed to fetch questions count: %v", err)
	}

	var mu sync.Mutex
	var allQuestionsRetrieved []scraper.Question

	// query the questions in batches (evenly distributed)
	batchSize := scraper.GetBatchSize(totalQuestionsCount, scraper.MAX_GO_ROUTINES)

	log.Printf("attempting to query %d questions in %d batches of size %d. [~%d req/s]\n", totalQuestionsCount, scraper.MAX_GO_ROUTINES, batchSize, scraper.MAX_GO_ROUTINES)

	err = scraper.Parallel{Concurrency: scraper.MAX_GO_ROUTINES, BatchSize: batchSize}.Process(ctx, totalQuestionsCount, func(start, end int) {
		offset, limit := start, end-start

		questions, err := client.FetchQuestions(offset, limit)

		log.Printf("fetched %d questions.\n", len(questions))

		if err != nil {
			log.Printf("fetch went wrong: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()
		allQuestionsRetrieved = append(allQuestionsRetrieved, questions...)

	})

	if err != nil {
		log.Fatalf("failed to fetch questions: %v", err)
	}

	log.Printf("fetched %d questions from LeetCode.\n", len(allQuestionsRetrieved))

	if len(allQuestionsRetrieved) == 0 {
		log.Println("nothing to retrieved, therefore nothing to write in output.")
		return
	}

	// same thing, batch the retrieved questions evenly for html file output writes
	batchSize = scraper.GetBatchSize(len(allQuestionsRetrieved), scraper.MAX_GO_ROUTINES)

	// just to keep track of status
	filesStatusCounter := scraper.FileStatusCounter{
		Success: 0,
		Failure: 0,
		Skipped: 0,
	}

	// make the file writes in parallel
	err = scraper.Parallel{Concurrency: scraper.MAX_GO_ROUTINES, BatchSize: batchSize}.Process(ctx, len(allQuestionsRetrieved), func(start, end int) {
		counter := scraper.WriteHtmlFilesBatch(allQuestionsRetrieved[start:end])

		mu.Lock()
		defer mu.Unlock()

		filesStatusCounter.Success += counter.Success
		filesStatusCounter.Failure += counter.Failure
		filesStatusCounter.Skipped += counter.Skipped
	})

	log.Printf("Successfully wrote %d out of %d files retrieved. %d were skipped and %d failed.\n", filesStatusCounter.Success, len(allQuestionsRetrieved), filesStatusCounter.Skipped, filesStatusCounter.Failure)
}
