package cmd

import (
	"context"

	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/open_search"
	"github.com/spf13/cobra"
)

var initOpensearchCommand = &cobra.Command{
	Use:   "init-opensearch",
	Short: "Create the OpenSearch models, pipelines & indexes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg := ctx.Value(ConfigKey{}).(config.Config)
		logger := logging.FromContext(ctx)

		logger.Info("Creating model")

		modelId := open_search.CreateModel(cfg, ctx, logger)
		cfg.ModelId = modelId

		logger.Info("Model created")

		ctx = context.WithValue(cmd.Context(), ConfigKey{}, cfg)
		cmd.SetContext(ctx)

		logger.Info("Loading model")

		if err := open_search.LoadModel(cfg, ctx, logger); err != nil {
			return err
		}

		logger.Info("Model loaded")

		open_search.CreatePipeline(cfg, ctx, logger)

		logger.Info("Pipeline created")

		open_search.CreateIndex(cfg, ctx, logger)

		logger.Info("Index created")

		return nil
	},
}

type ConfigKey struct{}
