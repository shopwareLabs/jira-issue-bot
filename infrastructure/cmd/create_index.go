package cmd

import (
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/open_search"
	"github.com/spf13/cobra"
)

var createIndexCommand = &cobra.Command{
	Use:   "create-index",
	Short: "Create an OpenSearch index with the currently configured name",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cfg := ctx.Value(ConfigKey{}).(config.Config)
		logger := logging.FromContext(ctx)

		open_search.CreateIndex(cfg, ctx, logger)
	},
}
