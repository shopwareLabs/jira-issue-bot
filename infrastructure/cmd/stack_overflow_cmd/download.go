package stack_overflow_cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/shopwarelabs/jira-issue-bot/domain/stack_overflow_connector"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var downloadStackOverflowCommand = &cobra.Command{
	Use:   "stack-overflow",
	Short: "Download issues from Stack Overflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.FromContext(cmd.Context())

		logger.Debug("Start downloading Stack Overflow questions")
		return extractQuestions(1, cmd.Context(), logger)
	},
}

func extractQuestions(page int, ctx context.Context, logger *zap.SugaredLogger) error {
	logger.Debugf("Downloads page %d of Stack Overflow questions", page)

	questions, err := stack_overflow_connector.GetQuestions(page, "creation", "shopware6", ctx)

	if err != nil {
		return err
	}

	if _, err := os.Stat("stack-overflow"); os.IsNotExist(err) {
		if err := os.Mkdir("stack-overflow", os.ModePerm); err != nil {
			return err
		}
	}

	for _, question := range questions.Items {
		data, _ := json.Marshal(question)
		if err := os.WriteFile(fmt.Sprintf("stack-overflow/%d.json", question.QuestionId), data, 0600); err != nil {
			return err
		}
	}

	if questions.HasMore {
		return extractQuestions(page+1, ctx, logger)
	}

	return nil
}
