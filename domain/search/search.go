package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
)

func Search(title string, description string, filter SearchFilter, config config.Config) (*SearchResponse, error) {
	encodedTitle, _ := json.Marshal(title)
	encodedDescription, _ := json.Marshal(CleanupString(description))

	search := parseFilter(string(encodedTitle), string(encodedDescription), config.ModelId, filter)

	req := opensearchapi.SearchRequest{
		Index: []string{config.IndexName},
		Body:  strings.NewReader(search.String()),
	}

	resp, err := req.Do(context.Background(), config.OpensearchClient)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	var result SearchResponse

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode opensearch response: %w", err)
	}

	return &result, nil
}

func SearchId(id string, filter SearchFilter, config config.Config) (*SearchResponse, error) {
	docReq := opensearchapi.GetRequest{
		Index:      config.IndexName,
		DocumentID: id,
	}

	docResp, err := docReq.Do(context.Background(), config.OpensearchClient)
	if err != nil {
		return nil, fmt.Errorf("failed to execute document fetch: %w", err)
	}

	var issue IssueSearchResponse

	docBody, _ := io.ReadAll(docResp.Body)
	defer docResp.Body.Close()

	if err := json.Unmarshal(docBody, &issue); err != nil {
		return nil, fmt.Errorf("failed to decode opensearch response: %w", err)
	}

	if !issue.Found {
		return nil, fmt.Errorf("issue not found")
	}

	return Search(issue.Source.Title, issue.Source.Description, filter, config)
}

type SearchFilter struct {
	ExcludedDocumentId string
	Source             string
	OnlyPublic         bool
}

type SearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64       `json:"max_score"`
		Hits     []IssueResult `json:"hits"`
	} `json:"hits"`
}

type IssueResult struct {
	Index  string   `json:"_index"`
	ID     string   `json:"_id"`
	Score  float64  `json:"_score"`
	Source Document `json:"_source"`
}

type IssueSearchResponse struct {
	Index  string   `json:"_index"`
	ID     string   `json:"_id"`
	Found  bool     `json:"found"`
	Source Document `json:"_source"`
}
