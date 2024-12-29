package scraper

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
)

func WriteHtmlFilesBatch(questions []Question) FileStatusCounter {
	counter := FileStatusCounter{
		Success: 0,
		Failure: 0,
		Skipped: 0,
	}

	for _, question := range questions {
		if question.Content == nil {
			// log.Printf("%s is a premium problem.\n", question.Id+". "+question.Title)
			counter.Skipped++
			continue
		}
		pathname := QUESTIONS_OUTPUT_DIR + question.Id + "." + question.TitleSlug + ".html"
		if _, err := os.Stat(pathname); err == nil {
			// log.Printf("%s exists already.\n", pathname)
			counter.Skipped++
			continue
		}

		file, err := os.Create(pathname)
		if err != nil {
			log.Printf("Error creating %s: %v\n", pathname, err)
			counter.Failure++
			continue
		}
		defer file.Close()

		writer := bufio.NewWriter(file)

		htmlFileContent := buildHtmlFile(question)
		_, err = writer.WriteString(htmlFileContent)
		if err != nil {
			log.Printf("Error writing string for %s: %v\n", pathname, err)
			counter.Failure++
			continue
		}

		err = writer.Flush()
		if err != nil {
			log.Printf("Error writing to the file %s: %v\n", pathname, err)
			counter.Failure++
			continue
		}
		counter.Success++
	}
	return counter
}

// TODO: segment this into html builder
func buildHtmlFile(question Question) string {
	var sb strings.Builder

	sb.WriteString("<title>" + question.Id + ". " + question.Title + "</title>\n")
	sb.WriteString("<strong>" + question.Difficulty + "</strong>\n")
	if len(question.Topics) > 0 {
		sb.WriteString("<ul>\n")
		sb.WriteString("<li><strong>Topics</strong></li>\n")
		for _, topic := range question.Topics {
			sb.WriteString("<li>\n")
			sb.WriteString("<p>" + topic.Name + "</p>\n")
			sb.WriteString("</li>\n")
		}
		sb.WriteString("</ul>\n")
	}

	var codeDefinitions []CodeDefinition
	json.Unmarshal([]byte(question.CodeDefinitions), &codeDefinitions)

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
	sb.WriteString(*question.Content)

	return sb.String()
}
