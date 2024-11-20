package github_cmd

import (
	"encoding/json"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/shopwarelabs/jira-issue-bot/domain/github_connector"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/cmd"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"
	"github.com/spf13/cobra"
)

var indexGithubCommand = &cobra.Command{
	Use:   "github",
	Short: "Index all downloaded github issues to OpenSearch",
	RunE: func(command *cobra.Command, args []string) error {
		cfg := command.Context().Value(cmd.ConfigKey{}).(config.Config)
		logger := logging.FromContext(command.Context())

		guard := make(chan struct{}, 14)

		files, _ := os.ReadDir("github")
		for _, file := range files {
			var issue github.Issue

			readFile, _ := os.ReadFile("github/" + file.Name())
			if err := json.Unmarshal(readFile, &issue); err != nil {
				logger.Error(err)
				continue
			}

			guard <- struct{}{}

			go func(issue *github.Issue) {
				if err := github_connector.IndexSingleGitHubIssue(issue, cfg, logger); err != nil {
					logger.Error(err)
				}
				<-guard
			}(&issue)
		}

		return nil
	},
}

func Register(rootCmd *cobra.Command, downloadCommand *cobra.Command, indexCommand *cobra.Command) {
	indexCommand.AddCommand(indexGithubCommand)
	downloadCommand.AddCommand(downloadGithubCommand)
}
