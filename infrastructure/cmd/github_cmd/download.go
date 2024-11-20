package github_cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/cmd"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/spf13/cobra"
)

var downloadGithubCommand = &cobra.Command{
	Use:   "github",
	Short: "Download issues from GitHub",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := command.Context()
		cfg := ctx.Value(cmd.ConfigKey{}).(config.Config)

		if _, err := os.Stat("github"); os.IsNotExist(err) {
			if err := os.Mkdir("github", os.ModePerm); err != nil {
				return err
			}
		}

		options := &github.IssueListByRepoOptions{
			ListOptions: github.ListOptions{PerPage: 100},
			State:       "all",
		}

		return extractGithubIssues(cfg.GithubClient, options, ctx)
	},
}

func extractGithubIssues(client *github.Client, options *github.IssueListByRepoOptions, ctx context.Context) error {
	issues, response, err := client.Issues.ListByRepo(ctx, "shopware", "platform", options)

	if err != nil {
		return err
	}

	for _, issue := range issues {
		data, _ := json.Marshal(issue)
		if err := os.WriteFile(fmt.Sprintf("github/issue-%d.json", issue.GetNumber()), data, 0600); err != nil {
			return err
		}
	}

	if response.NextPage == 0 {
		return nil
	}

	options.Page = response.NextPage

	return extractGithubIssues(client, options, ctx)
}
