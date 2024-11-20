package stack_overflow_cmd

import (
	"github.com/shopwarelabs/jira-issue-bot/domain/stack_overflow_connector"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/cmd"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"
	"github.com/spf13/cobra"
)

var cronCommand = &cobra.Command{
	Use:   "stack-overflow-cron",
	Short: "Reindex the latest stack overflow questions",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := command.Context()
		cfg := ctx.Value(cmd.ConfigKey{}).(config.Config)
		logger := logging.FromContext(ctx)

		// Fetch only the first page, should be enough as we execute this every hour
		questions, err := stack_overflow_connector.GetQuestions(1, "activity", "shopware6", ctx)
		if err != nil {
			return err
		}

		guard := make(chan struct{}, 14)

		for _, question := range questions.Items {
			guard <- struct{}{}

			go func(question stack_overflow_connector.StackoverflowListingElement) {
				if err := stack_overflow_connector.IndexSingleStackOverflowQuestion(&question, cfg, logger); err != nil {
					logger.Error(err)
				}
				<-guard
			}(question)
		}

		return nil
	},
}
