package slack_connector

import (
	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

func OnIssuesCommand(command slack.SlashCommand, config config.Config, logger *zap.SugaredLogger) (slack.Message, error) {
	logger.Debugf("Try to find recommendations for message: %s", command.TriggerID)

	result, err := search.Search(
		command.Text,
		command.Text,
		search.SearchFilter{},
		config,
	)

	if err != nil {
		return slack.Message{}, err
	}

	if len(result.Hits.Hits) == 0 {
		logger.Debugf("Did not found any recommendation for message: %s", command.TriggerID)
		return slack.NewBlockMessage(noResultsSectionBlock()), nil
	}

	logger.Debugf("Found %d recommendations for message %s", len(result.Hits.Hits), command.TriggerID)

	headerSection := headerSectionBlock()
	listSection := resultListSectionBlock(result)

	return slack.NewBlockMessage(headerSection, listSection), nil
}
