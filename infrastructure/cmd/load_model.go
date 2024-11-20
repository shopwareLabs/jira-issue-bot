package cmd

import (
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/open_search"
	"github.com/spf13/cobra"
)

var loadModelCommand = &cobra.Command{
	Use:   "load-model",
	Short: "Loads the ML model after a restart",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg := ctx.Value(ConfigKey{}).(config.Config)
		logger := logging.FromContext(ctx)

		return open_search.LoadModel(cfg, ctx, logger)
	},
}
