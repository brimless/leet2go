package scraper

import (
	"io"
	"math/rand/v2"
	"os"
	"strings"
)

// fetches user-agents from the user-agents.txt in root
func fetchUserAgents() []string {
	userAgents := []string{} // return an empty list if no user-agents found

	userAgentsFile, err := os.Open("user-agents.txt")
	if err != nil {
		return userAgents
	}
	defer userAgentsFile.Close()

	userAgentsBytes, err := io.ReadAll(userAgentsFile)
	if err != nil {
		return userAgents
	}

	userAgentsStr := string(userAgentsBytes)
	userAgentsStr = strings.TrimSuffix(strings.ReplaceAll(userAgentsStr, "\r", "\n"), "\n")

	if len(userAgentsStr) == 0 {
		return userAgents
	}

	for _, ua := range strings.Split(userAgentsStr, "\n") {
		if ua != "" {
			userAgents = append(userAgents, ua)
		}
	}

	return userAgents
}

func chooseRandomUserAgent(userAgents []string) string {
	return userAgents[rand.IntN(len(userAgents))]
}
