package scraper

import (
	"net/http"
	"net/url"
)

// LEETCODE
type LeetCodeResponse struct {
	Data struct {
		ProblemsetQuestionList struct {
			Total     int        `json:"total"`
			Questions []Question `json:"questions"`
		} `json:"problemsetQuestionList"`
	} `json:"data"`
}

type Question struct {
	Id              string          `json:"frontendQuestionId"`
	TitleSlug       string          `json:"titleSlug"`
	Title           string          `json:"title"`
	Difficulty      string          `json:"difficulty"`
	AcceptanceRate  float64         `json:"acRate"`
	Hints           []string        `json:"hints"`
	Topics          []QuestionTopic `json:"topicTags"`
	CodeDefinitions string          `json:"codeDefinition"` // it's actually an array of objects, but it comes in as a string
	Content         *string         `json:"content"`        // this is null only if it is a premium problem
}

type QuestionTopic struct {
	Name string `json:"name"`
}

type CodeDefinition struct {
	Value       string `json:"value"`
	Text        string `json:"text"`
	DefaultCode string `json:"defaultCode"`
}

type LeetCodeClient struct {
	Client     *http.Client
	UserAgents []string
}

type LeetCodePayloadVariables struct {
	CategorySlug string                 `json:"categorySlug"`
	Filters      map[string]interface{} `json:"filters"`
	Limit        int                    `json:"limit"`
	Skip         int                    `json:"skip"`
}

type LeetCodePayload struct {
	Query         string                   `json:"query"`
	Variables     LeetCodePayloadVariables `json:"variables"`
	OperationName string                   `json:"operationName"`
}

// OUTPUT FILES
type FileStatusCounter struct {
	Success int
	Failure int
	Skipped int
}

// PROXIES
type RandomProxySelector struct {
	Proxies []*url.URL
}
