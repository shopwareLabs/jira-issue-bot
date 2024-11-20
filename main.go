package main

import (
	"context"
	"fmt"
	"os"

	"github.com/shopwarelabs/jira-issue-bot/infrastructure/cmd"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/cmd/github_cmd"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/cmd/stack_overflow_cmd"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/subosito/gotenv"
)

var rootCmd = &cobra.Command{
	Use: "issue-dedupe",
}

var downloadCommand = &cobra.Command{
	Use:   "download",
	Short: "Download issues from Platforms",
}

var indexCommand = &cobra.Command{
	Use:   "index",
	Short: "Index issues from Platforms",
}

func main() {
	if fileExists(".env") {
		_ = gotenv.Load(".env")
	} else if fileExists(".env.dist") {
		_ = gotenv.Load(".env.dist")
	}

	ctx := rootCmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	cfg, err := config.NewFromEnv(ctx)
	if err != nil {
		panic(fmt.Sprintf("Error loading config from env: %v\n", err))
	}

	logger := config.NewLogger(cfg)

	defer logger.Sync() //nolint:errcheck
	ctx = logging.WithLogger(ctx, logger)

	rootCmd.SetContext(context.WithValue(ctx, cmd.ConfigKey{}, cfg))

	rootCmd.AddCommand(downloadCommand)
	rootCmd.AddCommand(indexCommand)
	rootCmd.AddCommand(serverCommand)
	cmd.Register(rootCmd)
	github_cmd.Register(rootCmd, downloadCommand, indexCommand)
	stack_overflow_cmd.Register(rootCmd, downloadCommand, indexCommand)

	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
