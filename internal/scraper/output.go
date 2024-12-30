package scraper

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
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

		err = parseHtmlFile(question, file)
		if err != nil {
			log.Printf("Error while parsing template %s: %v\n", pathname, err)
			counter.Failure++
			continue
		}

		counter.Success++
	}
	return counter
}

func parseHtmlFile(question Question, file *os.File) error {
	funcs := template.FuncMap{
		"addInt":            AddInt,
		"addFloat":          AddFloat,
		"mult":              Multiply,
		"float":             ToFloat,
		"getQuestionColour": GetQuestionDifficultyColour,
	}

	t, err := template.New("template.html").Funcs(funcs).ParseFiles("template/template.html")
	if err != nil {
		return err
	}

	var codeDefinitions []CodeDefinition
	json.Unmarshal([]byte(question.CodeDefinitions), &codeDefinitions)

	data := map[string]interface{}{
		"Id":              question.Id,
		"Title":           question.Title,
		"Difficulty":      question.Difficulty,
		"AcceptanceRate":  fmt.Sprintf("%.2f%%", question.AcceptanceRate),
		"Hints":           question.Hints,
		"Topics":          question.Topics,
		"Content":         template.HTML(*question.Content),
		"CodeDefinitions": codeDefinitions,
		"DefaultLang":     DEFAULT_LANG,
	}

	err = t.Execute(file, data)
	return err
}
