package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
)

func IndexDocument(id string, document Document, config config.Config) error {
	if document.Description == "" {
		document.Description = document.Title
	}

	jsonString, _ := json.Marshal(document)

	req := opensearchapi.IndexRequest{
		Index:      config.IndexName,
		DocumentID: id,
		Body:       bytes.NewReader(jsonString),
	}

	resp, err := req.Do(context.Background(), config.OpensearchClient)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	defer resp.Body.Close()

	if resp.IsError() {
		body, er := io.ReadAll(resp.Body)
		if er != nil {
			return fmt.Errorf("failed to read response body: %w", er)
		}

		return fmt.Errorf("failed to insert document: %s", body)
	}

	return nil
}

type Document struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Type         string   `json:"type"`
	Link         string   `json:"link"`
	ExternalLink string   `json:"externalLink"`
	FixVersion   []string `json:"fixVersion"`
	Public       bool     `json:"public"`
	Source       string   `json:"source"`
	AuthorName   string   `json:"authorName"`
	AuthorLink   string   `json:"authorLink"`
	DateCreated  int64    `json:"dateCreated"`
	Labels       []string `json:"labels"`
}
