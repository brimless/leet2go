package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// TODO: segment files

const MAX_GO_ROUTINES = 10
const MAX_RETRY = 10
const PROXIES_API_URL = "https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=20000&country=us,ca&ssl=all&anonymity=elite"

// proxies should in theory protect against IP bans, but it seems like they're not reliable (at least the free ones)
// anyways, for this leetcode in specific, it probably doesn't matter since we don't need that many queries
const USE_PROXY = false // TODO: make this a param

const LEETCODE_URL = "https://leetcode.com/graphql/"

const QUESTION_OUTPUT_DIR = "problems/"

var availableProxies []string

type CodeDefinition struct {
	Value       string `json:"value"`
	Text        string `json:"text"`
	DefaultCode string `json:"defaultCode"`
}

type Question struct {
	Id         string   `json:"frontendQuestionId"`
	TitleSlug  string   `json:"titleSlug"`
	Title      string   `json:"title"`
	Difficulty string   `json:"difficulty"`
	Hints      []string `json:"hints"`
	Topics     []struct {
		Name string `json:"name"`
	} `json:"topicTags"`
	CodeDefinitions string  `json:"codeDefinition"` // it's actually an array of objects, but it comes in as a string
	Content         *string `json:"content"`        // this is null only if it is a premium problem
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
	if USE_PROXY {
		// fetch some random proxies
		proxies, err := fetchProxies()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Found %d proxies.\n", len(proxies))
		availableProxies = proxies
	}

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

	var httpClient *http.Client
	if USE_PROXY {
		httpClient = &http.Client{
			Transport: &http.Transport{Proxy: chooseRandomProxy}, // choose a random proxy given a list of proxies
		}
	} else {
		httpClient = &http.Client{}
	}

	questionsChannel := make(chan []Question)
	allQuestions := []Question{}

	go func() {
		wg.Wait()
		defer close(questionsChannel)
	}()

	for totalProblemsCount > 0 {
		if problemCountPerQuery > totalProblemsCount {
			problemCountPerQuery = totalProblemsCount
		}

		wg.Add(1)
		// TODO: also implement caching
		go func(skip, limit int, c chan<- []Question) {
			// signal that we are done with this request
			defer wg.Done()

			// do request using a random user agent
			userAgent := userAgents[rand.IntN(len(userAgents))]
			// TODO: this should return an error, and we should send it to a channel or smth
			makeHttpRequests(skip, limit, c, userAgent, httpClient)
		}(offset, problemCountPerQuery, questionsChannel)

		offset += problemCountPerQuery
		totalProblemsCount -= problemCountPerQuery
	}

	for questions := range questionsChannel {
		allQuestions = append(allQuestions, questions...)
	}

	fmt.Printf("Total Questions Fetched: %d\n", len(allQuestions))

	if len(allQuestions) == 0 {
		fmt.Println("Nothing to write.")
		return
	}

	// write files based questions results array
	// maybe use go routines as well
	writeWg := sync.WaitGroup{}
	offset = 0
	questionsCount := len(allQuestions)
	for questionsCount > 0 {
		if problemCountPerQuery > questionsCount {
			problemCountPerQuery = questionsCount
		}
		writeWg.Add(1)
		go func(skip, limit int) {
			defer writeWg.Done()
			// TODO: this should return an error, and we should send it to a channel or smth
			writeHtmlFiles(skip, limit, allQuestions)
		}(offset, problemCountPerQuery)

		offset += problemCountPerQuery
		questionsCount -= problemCountPerQuery
	}
	writeWg.Wait()

	fmt.Println("Done!")
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

	proxies := strings.Split(strings.ReplaceAll(proxiesBodyStr, "\r", "\n"), "\n")
	var cleanedProxies []string
	for i := range proxies {
		if proxies[i] != "" {
			cleanedProxies = append(cleanedProxies, "http://"+proxies[i]) // they should all be http proxies
		}
	}

	return cleanedProxies, nil
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

func chooseRandomProxy(req *http.Request) (*url.URL, error) {
	proxy := availableProxies[rand.IntN(len(availableProxies))]
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy url %s: %v", proxyUrl, err)
	}
	fmt.Printf("using proxy: %s for request to %s\n", proxy, req.URL)
	return proxyUrl, nil
}

// TODO: probably return error
func makeHttpRequests(skip, limit int, c chan<- []Question, userAgent string, httpClient *http.Client) {
	// choose a random proxy and user agent from the options

	fmt.Printf("doing request with skip: %d, limit: %d, user-agent: %s\n", skip, limit, userAgent)

	payload := map[string]interface{}{
		"query": `
			query problemsetQuestionList($categorySlug:String,$limit:Int,$skip:Int,$filters:QuestionListFilterInput){
				problemsetQuestionList:questionList(
					categorySlug:$categorySlug
					limit:$limit
					skip:$skip
					filters:$filters
				){
					questions:data{
						frontendQuestionId:questionFrontendId
						titleSlug
						title
						difficulty
						hints
						topicTags{name}
						codeDefinition
						content
					}
				}
			}`,
		"variables": map[string]interface{}{
			"categorySlug": "all-code-essentials",
			"skip":         skip,
			"limit":        limit,
			"filters":      map[string]interface{}{},
		},
		"operationName": "problemsetQuestionList",
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		// TODO: figure out error handling
		log.Fatalf("failed to marshal payload: %v", err)
	}
	req, err := http.NewRequest("POST", LEETCODE_URL, bytes.NewBuffer(jsonBody))

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	retryCount := 1
	var resp *http.Response

	// TODO: actually do good error handling when proxies fail
	// should probably return a status or whatever somewhere
	for retryCount <= MAX_RETRY {
		if retryCount == MAX_RETRY {
			log.Fatalf("failed after %d retries: %v", retryCount, err)
		}
		resp, err = httpClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Printf("failed to do the request at skip: %d and limit: %d: %v", skip, limit, err)
			retryCount += 1
			continue
		}
		break
	}
	defer resp.Body.Close()

	// read from response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// TODO: figure out error handling
		log.Fatalf("failed to read body of response: %v", err)
	}
	var respObj LeetCodeResponse
	json.Unmarshal(respBody, &respObj)

	log.Printf("Found %d questions.\n", len(respObj.Data.ProblemsetQuestionList.Questions))
	c <- respObj.Data.ProblemsetQuestionList.Questions
}

// TODO: format html file better
// TODO: probably return error
func writeHtmlFiles(skip, limit int, allQuestions []Question) {
	for i := skip; i-skip < limit; i++ {
		curr := allQuestions[i]
		if curr.Content == nil {
			log.Printf("%s is a premium problem.\n", curr.Id+". "+curr.Title)
			continue
		}
		pathname := QUESTION_OUTPUT_DIR + curr.Id + "." + curr.TitleSlug + ".html"
		if _, err := os.Stat(pathname); err == nil {
			log.Printf("%s exists already.\n", pathname)
			continue
		}

		file, err := os.Create(pathname)
		if err != nil {
			log.Printf("Error creating %s: %v\n", pathname, err)
			continue
		}
		defer file.Close()

		var sb strings.Builder
		sb.WriteString("<title>" + curr.Id + ". " + curr.Title + "</title>\n")
		sb.WriteString("<strong>" + curr.Difficulty + "</strong>\n")
		if len(curr.Topics) > 0 {
			sb.WriteString("<ul>\n")
			sb.WriteString("<li><strong>Topics</strong></li>\n")
			for _, topic := range curr.Topics {
				sb.WriteString("<li>\n")
				sb.WriteString("<p>" + topic.Name + "</p>\n")
				sb.WriteString("</li>\n")
			}
			sb.WriteString("</ul>\n")
		}

		var codeDefinitions []CodeDefinition
		json.Unmarshal([]byte(curr.CodeDefinitions), &codeDefinitions)

		sb.WriteString("<ul>\n")
		for _, codeDefinition := range codeDefinitions {
			sb.WriteString("<li>\n")
			sb.WriteString("<p>" + codeDefinition.Text + "</p>\n")
			sb.WriteString("<code><pre>\n")
			sb.WriteString(codeDefinition.DefaultCode)
			sb.WriteString("</pre></code>\n")
			sb.WriteString("</li>\n")
		}
		sb.WriteString("</ul>\n")
		sb.WriteString(*curr.Content)

		writer := bufio.NewWriter(file)
		// TODO: also add the other elements (difficulty, title, etc...)
		_, err = writer.WriteString(sb.String())
		if err != nil {
			log.Printf("Error writing to %s: %v\n", pathname, err)
			continue
		}

		err = writer.Flush()
		if err != nil {
			log.Printf("Error flushing writer: %v\n", err)
			continue
		}
	}
}
