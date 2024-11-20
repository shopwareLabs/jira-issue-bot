package github_connector

import (
	"fmt"

	"github.com/google/go-github/v50/github"
	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"go.uber.org/zap"
)

func IndexSingleGitHubIssue(issue *github.Issue, config config.Config, logger *zap.SugaredLogger) error {
	var issueType string
	if issue.IsPullRequest() {
		issueType = "Pull Request"
	} else {
		issueType = "Issue"
	}

	var labels []string
	for _, label := range issue.Labels {
		labels = append(labels, label.GetName())
	}

	document := search.Document{
		Title:        issue.GetTitle(),
		Description:  search.CleanupString(issue.GetBody()),
		Status:       issue.GetState(),
		Type:         issueType,
		Link:         issue.GetHTMLURL(),
		ExternalLink: issue.GetHTMLURL(),
		FixVersion:   []string{"n/a"},
		Public:       true,
		Source:       "github",
		AuthorName:   issue.GetUser().GetLogin(),
		AuthorLink:   issue.GetUser().GetHTMLURL(),
		DateCreated:  issue.CreatedAt.Unix(),
		Labels:       labels,
	}

	err := search.IndexDocument(fmt.Sprintf("GH-%d", issue.GetNumber()), document, config)
	if err != nil {
		return err
	}

	logger.Debugf("Indexed GitHub issue: %d", issue.GetNumber())

	return nil
}

func IndexSingleGitHubPr(pr *github.PullRequest, config config.Config, logger *zap.SugaredLogger) error {
	var labels []string
	for _, label := range pr.Labels {
		labels = append(labels, label.GetName())
	}

	document := search.Document{
		Title:        pr.GetTitle(),
		Description:  search.CleanupString(pr.GetBody()),
		Status:       pr.GetState(),
		Type:         "Pull Request",
		Link:         pr.GetHTMLURL(),
		ExternalLink: pr.GetHTMLURL(),
		FixVersion:   []string{"n/a"},
		Public:       true,
		Source:       "github",
		AuthorName:   pr.GetUser().GetLogin(),
		AuthorLink:   pr.GetUser().GetHTMLURL(),
		DateCreated:  pr.CreatedAt.Unix(),
		Labels:       labels,
	}

	err := search.IndexDocument(fmt.Sprintf("GH-%d", pr.GetNumber()), document, config)
	if err != nil {
		return err
	}

	logger.Debugf("Indexed GitHub pull request: %d", pr.GetNumber())

	return nil
}
