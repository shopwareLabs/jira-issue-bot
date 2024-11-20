package github_connector

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"go.uber.org/zap"
)

func HandleGithubIssueEvent(event *github.IssuesEvent, config config.Config, logger *zap.SugaredLogger) error {
	if err := IndexSingleGitHubIssue(event.GetIssue(), config, logger); err != nil {
		logger.Errorf("Error while indexing GitHub issue: %s", err)
	}

	if event.GetAction() != "opened" {
		return nil
	}

	result, err := search.Search(
		event.GetIssue().GetTitle(),
		event.GetIssue().GetBody(),
		search.SearchFilter{ExcludedDocumentId: fmt.Sprintf("GH-%d", event.GetIssue().GetNumber()), OnlyPublic: true},
		config,
	)

	if err != nil {
		return err
	}

	if len(result.Hits.Hits) == 0 {
		logger.Debugf("Did not find any recommendations for issue: GH-%d", event.GetIssue().GetNumber())
		return nil
	}

	var output strings.Builder
	output.WriteString("We found the following existing issues which may help or are related to your topic: \n")

	logger.Debugf("Found %d recommendations for ticket GH-%d", len(result.Hits.Hits), event.GetIssue().GetNumber())

	for _, hit := range result.Hits.Hits {
		switch hit.Source.Source {
		case "jira":
			output.WriteString(fmt.Sprintf("- [%s](%s)\n", hit.Source.Title, "https://issues.shopware.com/issues/"+hit.ID))
		case "github":
			link := "https://github.com/shopware/platform/issues/" + strings.Replace(hit.ID, "GH-", "", 1)
			output.WriteString(fmt.Sprintf("- [%s](%s)\n", hit.Source.Title, link))
		}
	}

	message := output.String()
	_, _, err = config.GithubClient.Issues.CreateComment(context.Background(), "shopware", "platform", event.GetIssue().GetNumber(), &github.IssueComment{Body: &message})

	return err
}

func HandleGithubPREvent(event *github.PullRequestEvent, config config.Config, logger *zap.SugaredLogger) error {
	if err := IndexSingleGitHubPr(event.GetPullRequest(), config, logger); err != nil {
		logger.Errorf("Error while indexing GitHub pull request: %s", err)

		return err
	}

	return nil
}
