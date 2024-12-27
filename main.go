package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"sync"
)

const MAX_GO_ROUTINES = 10
const PROXIES_API_URL = "https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=anonymous"

const LEETCODE_URL = "https://leetcode.com/graphql/"

type Question struct {
	Id         string   `json:"frontendQuestionId"`
	TitleSlug  string   `json:"titleSlug"`
	Title      string   `json:"title"`
	Difficulty string   `json:"difficulty"`
	Hints      []string `json:"hints"`
	Topics     []struct {
		Name string `json:"name"`
	} `json:"topicTags"`
	CodeDefinition string  `json:"codeDefinition"`
	Content        *string `json:"content"` // this is null only if it is a premium problem
}

type LeetCodeResponse struct {
	Data struct {
		ProblemsetQuestionList struct {
			Total     int        `json:"total"`
			Questions []Question `json:"questions"`
		} `json:"problemsetQuestionList"`
	} `json:"data"`
}

func main() {
	// fetch some random proxies
	proxies, err := fetchProxies()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d proxies.\n", len(proxies))

	// fetch user-agents from local user-agents.txt file
	userAgents, err := fetchUserAgents()

	fmt.Printf("Found %d user-agents.\n", len(userAgents))

	// fetch total number of problems on leetcode
	// NOTE: i think it's fine if we don't spoof our client here since it's only 1 request?
	totalProblemsCount, err := fetchProblemCount()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("There are %d problems on LeetCode currently.\n", totalProblemsCount)

	// TODO: implement some kind of caching to avoid querying
	// this might be hard because we can only query by chunks
	// maybe if we have the same amount of files as the problems count, then we can ignore?
	// or we could do the query regardless, but just avoid the file write if the file already exists

	// divide up the requests given MAX_GO_ROUTINES
	problemCountPerQuery := int(math.Ceil(float64(totalProblemsCount) / MAX_GO_ROUTINES))
	fmt.Printf("problems per query: %v\n", problemCountPerQuery)

	wg := sync.WaitGroup{}
	offset := 0
	for totalProblemsCount > 0 {
		if problemCountPerQuery > totalProblemsCount {
			problemCountPerQuery = totalProblemsCount
		}

		wg.Add(1)
		// TODO: actually do the request here and store information somewhere
		// TODO: also implement caching
		go func(skip, limit int) {
			// signal that we are done with this request
			defer wg.Done()
			// choose a random proxy and user agent from the options
			// do request using that proxy and user agent combo
			userAgent := userAgents[rand.IntN(len(userAgents))]
			proxy := proxies[rand.IntN(len(proxies))]
			fmt.Printf("doing request with skip: %d, limit: %d, user-agent: %s, proxy: %s\n", skip, limit, userAgent, proxy)

			// read from response body

			// write files based on response
		}(offset, problemCountPerQuery)

		offset += problemCountPerQuery
		totalProblemsCount -= problemCountPerQuery

	}

	wg.Wait()

}

func fetchProxies() ([]string, error) {
	proxiesResp, err := http.Get(PROXIES_API_URL)
	if err != nil {
		return nil, fmt.Errorf("fetch proxies error: %v", err)
	}

	proxiesBody, err := io.ReadAll(proxiesResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read proxies response body error: %v", err)
	}
	defer proxiesResp.Body.Close()

	proxiesBodyStr := string(proxiesBody)
	proxiesBodyStr = strings.TrimSuffix(proxiesBodyStr, "\n")

	if len(proxiesBodyStr) == 0 {
		return nil, fmt.Errorf("found no proxies")
	}

	proxies := strings.Split(proxiesBodyStr, "\n")

	return proxies, nil
}

func fetchUserAgents() ([]string, error) {
	userAgentsFile, err := os.Open("user-agents.txt")
	if err != nil {
		return nil, fmt.Errorf("user-agents file doesn't exist: %v", err)
	}
	defer userAgentsFile.Close()

	userAgentsBytes, err := io.ReadAll(userAgentsFile)
	if err != nil {
		return nil, fmt.Errorf("read user-agents file error: %v", err)
	}

	userAgentsStr := string(userAgentsBytes)
	userAgentsStr = strings.TrimSuffix(userAgentsStr, "\n")

	if len(userAgentsStr) == 0 {
		return nil, fmt.Errorf("found no user-agents")
	}

	userAgents := strings.Split(userAgentsStr, "\n")
	return userAgents, nil
}

func fetchProblemCount() (int, error) {
	payload := map[string]interface{}{
		"query": "query problemsetQuestionList($categorySlug:String,$filters:QuestionListFilterInput){problemsetQuestionList:questionList(categorySlug:$categorySlug filters:$filters) {total:totalNum}}",
		"variables": map[string]interface{}{
			"categorySlug": "all-code-essentials",
			"filters":      map[string]interface{}{},
		},
		"operationName": "problemsetQuestionList",
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return -1, fmt.Errorf("error marshaling total number of problems request: %v", err)
	}

	totalProblemsResp, err := http.Post(LEETCODE_URL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return -1, fmt.Errorf("error fetching total problems count: %v", err)
	}
	defer totalProblemsResp.Body.Close()

	totalProblemsRespBody, err := io.ReadAll(totalProblemsResp.Body)
	if err != nil {
		return -1, fmt.Errorf("read total problems response body error: %v", err)
	}

	var totalProblemsObj LeetCodeResponse
	json.Unmarshal(totalProblemsRespBody, &totalProblemsObj)

	totalProblemsCount := totalProblemsObj.Data.ProblemsetQuestionList.Total
	return totalProblemsCount, nil
}
