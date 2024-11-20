package search

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func parseFilter(title string, description string, modelId string, filter SearchFilter) bytes.Buffer {
	must := make([]string, 0)
	mustNot := make([]string, 0)
	should := make([]string, 0)

	if filter.ExcludedDocumentId != "" {
		mustNot = append(mustNot, fmt.Sprintf(`{ "ids": { "values": ["%s"] }}`, filter.ExcludedDocumentId))
	}

	if filter.OnlyPublic {
		must = append(must, `{ "match": { "public": true }}`)
	}

	if filter.Source != "" {
		must = append(must, fmt.Sprintf(`{ "match": { "source": "%s" }}`, filter.Source))
	}

	should = append(should, fmt.Sprintf(`
{
  "script_score": {
    "query": {
      "neural": {
        "title_embedding": { "model_id": "%s", "k": 100, "query_text": %s }
      }
    },
    "script": { "source": "_score * 1.8"}
  }
}
`, modelId, title))

	should = append(should, fmt.Sprintf(`
{
  "script_score": {
    "query": {
      "neural": {
        "description_embedding": { "model_id": "%s", "k": 100, "query_text": %s }
      }
    },
    "script": { "source": "_score * 1.5" }
  }
}`, modelId, description))

	query := `
{
    "min_score": 1.8,
    "query": {
        "bool": {
            {{if .MustNot}}
            "must_not": [
                {{.MustNotString}}
            ],
            {{end}}
            {{if .Must}}
            "must": [
                {{.MustString}}
            ],
            {{end}}
            "should": [
                {{.ShouldString}}
            ]
        }
    }
}`

	tmpl, err := template.New("search").Parse(query)
	if err != nil {
		panic(err)
	}

	var search bytes.Buffer

	searchData := searchData{
		Must:          len(must) > 0,
		MustNot:       len(mustNot) > 0,
		Should:        len(should) > 0,
		MustString:    strings.Join(must, ","),
		MustNotString: strings.Join(mustNot, ","),
		ShouldString:  strings.Join(should, ","),
	}

	err = tmpl.Execute(&search, searchData)
	if err != nil {
		panic(err)
	}

	return search
}

type searchData struct {
	Must          bool
	MustNot       bool
	Should        bool
	MustString    string
	MustNotString string
	ShouldString  string
}
