package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func NewLeetCodeClient(useProxy bool) *LeetCodeClient {
	var httpClient *http.Client
	httpClient = &http.Client{
		Timeout: 5 * time.Second, // TODO: play with this value (also need to change the proxy api query to match this)
	}
	if useProxy {
		randomProxyTransport := getRandomProxyTransport()
		httpClient.Transport = randomProxyTransport
	}

	userAgents := fetchUserAgents()
	if len(userAgents) == 0 {
		// if there are no user-agents, we'll just default to a default user-agent
		userAgents = append(userAgents, DEFAULT_USER_AGENT)
	}

	return &LeetCodeClient{
		Client:     httpClient,
		UserAgents: userAgents,
	}
}

func NewLeetCodePayload(input ...int) *LeetCodePayload {
	var payload *LeetCodePayload
	payload = &LeetCodePayload{}
	payload.OperationName = "problemsetQuestionList"
	payload.Variables = LeetCodePayloadVariables{
		CategorySlug: "all-code-essentials",
		Filters:      map[string]interface{}{}, // this seems to always be {}
	}

	switch len(input) {
	case 0:
		payload.Query = QUESTIONS_COUNT_QUERY_STRING
	case 2:
		payload.Query = QUESTIONS_QUERY_STRING
		payload.Variables.Skip = input[0]
		payload.Variables.Limit = input[1]
	default:
		return nil
	}

	return payload
}

func (c *LeetCodeClient) FetchQuestionsCount() (int, error) {
	payload := NewLeetCodePayload()

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return -1, fmt.Errorf("error marshaling total number of questions request: %v", err)
	}

	req, err := http.NewRequest("POST", LEETCODE_URL, bytes.NewBuffer(jsonBody))

	userAgent := chooseRandomUserAgent(c.UserAgents)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	client := c.Client
	remainingRetries := MAX_RETRY
	var resp *http.Response

	for remainingRetries > 0 {
		err = nil
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			remainingRetries--
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}
		break
	}

	if err != nil {
		return -1, fmt.Errorf("error fetching total questions count: %v", err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("read total questions response body error: %v", err)
	}

	var respObj LeetCodeResponse
	json.Unmarshal(respBody, &respObj)

	return respObj.Data.ProblemsetQuestionList.Total, nil
}

func (c *LeetCodeClient) FetchQuestions(offset, limit int) ([]Question, error) {
	payload := NewLeetCodePayload(offset, limit)

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling questions request: %v", err)

	}
	req, err := http.NewRequest("POST", LEETCODE_URL, bytes.NewBuffer(jsonBody))

	userAgent := chooseRandomUserAgent(c.UserAgents)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	remainingRetries := MAX_RETRY
	var resp *http.Response

	client := c.Client
	for remainingRetries > 0 {
		err = nil // reset error
		resp, err = client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			remainingRetries--
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}
		break
	}

	// if there's still an error, then we know we have exhausted the retries
	if err != nil {
		return nil, fmt.Errorf("error making questions request after %d attempts: %v", MAX_RETRY, err)
	}

	defer resp.Body.Close()

	// read from response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read questions response body error: %v", err)
	}
	var respObj LeetCodeResponse
	json.Unmarshal(respBody, &respObj)

	return respObj.Data.ProblemsetQuestionList.Questions, nil
}
