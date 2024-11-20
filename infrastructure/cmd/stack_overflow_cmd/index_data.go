package stack_overflow_cmd

import (
	"encoding/json"
	"os"

	"github.com/shopwarelabs/jira-issue-bot/domain/stack_overflow_connector"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/cmd"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"
	"github.com/spf13/cobra"
)

var indexStackOverflowCommand = &cobra.Command{
	Use:   "stack-overflow",
	Short: "Index all downloaded StackOverflow questions to OpenSearch",
	RunE: func(command *cobra.Command, args []string) error {
		cfg := command.Context().Value(cmd.ConfigKey{}).(config.Config)
		logger := logging.FromContext(command.Context())

		guard := make(chan struct{}, 14)

		files, _ := os.ReadDir("stack-overflow")
		for _, file := range files {
			var question stack_overflow_connector.StackoverflowListingElement

			readFile, _ := os.ReadFile("stack-overflow/" + file.Name())
			if err := json.Unmarshal(readFile, &question); err != nil {
				return err
			}

			guard <- struct{}{}

			go func(question *stack_overflow_connector.StackoverflowListingElement) {
				if err := stack_overflow_connector.IndexSingleStackOverflowQuestion(question, cfg, logger); err != nil {
					logger.Error(err)
				}
				<-guard
			}(&question)
		}

		return nil
	},
}

func Register(rootCmd *cobra.Command, downloadCommand *cobra.Command, indexCommand *cobra.Command) {
	indexCommand.AddCommand(indexStackOverflowCommand)
	rootCmd.AddCommand(cronCommand)
	downloadCommand.AddCommand(downloadStackOverflowCommand)
}
