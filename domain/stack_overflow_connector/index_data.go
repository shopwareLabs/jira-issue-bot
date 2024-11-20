package stack_overflow_connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"go.uber.org/zap"
)

func GetQuestions(page int, sorting string, tag string, ctx context.Context) (*StackoverflowListingCollection, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.stackexchange.com/2.3/questions?page=%d&order=desc&sort=%s&tagged=%s&site=stackoverflow&filter=!nOedRLb*F(&pagesize=100", page, sorting, tag), nil)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("rate limited")
	}

	var collection StackoverflowListingCollection

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, err
	}

	return &collection, nil
}

func IndexSingleStackOverflowQuestion(question *StackoverflowListingElement, config config.Config, logger *zap.SugaredLogger) error {
	state := "open"
	if question.IsAnswered {
		state = "closed"
	}

	document := search.Document{
		Title:        question.Title,
		Description:  search.CleanupString(question.Body),
		Status:       state,
		Type:         "Question",
		Link:         question.Link,
		ExternalLink: question.Link,
		FixVersion:   []string{"n/a"},
		Public:       true,
		Source:       "stack-overflow",
		AuthorName:   question.Owner.DisplayName,
		AuthorLink:   question.Owner.Link,
		DateCreated:  question.CreationDate,
		Labels:       question.Tags,
	}

	err := search.IndexDocument(fmt.Sprintf("SO-%d", question.QuestionId), document, config)
	if err != nil {
		return err
	}

	logger.Debugf("Indexed StackOverflow question: %d", question.QuestionId)

	return nil
}

type StackoverflowListingCollection struct {
	Items          []StackoverflowListingElement `json:"items"`
	HasMore        bool                          `json:"has_more"`
	QuotaMax       int                           `json:"quota_max"`
	QuotaRemaining int                           `json:"quota_remaining"`
}

type StackoverflowListingElement struct {
	Tags             []string           `json:"tags"`
	Owner            StackoverflowOwner `json:"owner"`
	IsAnswered       bool               `json:"is_answered"`
	ViewCount        int                `json:"view_count"`
	ClosedDate       *int64             `json:"closed_date,omitempty"`
	AnswerCount      int                `json:"answer_count"`
	Score            int                `json:"score"`
	LastActivityDate int64              `json:"last_activity_date"`
	CreationDate     int64              `json:"creation_date"`
	LastEditDate     *int64             `json:"last_edit_date,omitempty"`
	QuestionId       int64              `json:"question_id"`
	Link             string             `json:"link"`
	ClosedReason     string             `json:"closed_reason,omitempty"`
	Title            string             `json:"title"`
	ContentLicense   string             `json:"content_license,omitempty"`
	AcceptedAnswerId *int               `json:"accepted_answer_id,omitempty"`
	Body             string             `json:"body_markdown"`
}

type StackoverflowOwner struct {
	AccountId    int    `json:"account_id"`
	Reputation   int    `json:"reputation"`
	UserId       int    `json:"user_id"`
	UserType     string `json:"user_type"`
	ProfileImage string `json:"profile_image"`
	DisplayName  string `json:"display_name"`
	Link         string `json:"link"`
}
