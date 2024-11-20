package slack_connector

import (
	"regexp"

	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.uber.org/zap"
)

func OnMention(event *slackevents.AppMentionEvent, config config.Config, logger *zap.SugaredLogger) error {
	api := slack.New(config.SlackBotToken)

	logger.Debugf("Try to find recommendations for message: %s", event.TimeStamp)

	regex := regexp.MustCompile(`<@U\d+[A-Z]+>`)
	searchTerm := regex.ReplaceAllString(event.Text, "")

	result, err := search.Search(
		searchTerm,
		searchTerm,
		search.SearchFilter{},
		config,
	)

	if err != nil {
		return err
	}

	if len(result.Hits.Hits) == 0 {
		logger.Debugf("Did not find any recommendations for message: %s", event.TimeStamp)

		_, _, err := api.PostMessage(
			event.Channel,
			slack.MsgOptionBlocks(noResultsSectionBlock()),
			slack.MsgOptionAsUser(true),
			slack.MsgOptionTS(event.TimeStamp),
		)

		if err != nil {
			return err
		}

		return nil
	}

	logger.Debugf("Found %d recommendations for message %s", len(result.Hits.Hits), event.TimeStamp)

	headerSection := headerSectionBlock()
	listSection := resultListSectionBlock(result)

	_, _, err = api.PostMessage(
		event.Channel,
		slack.MsgOptionBlocks(headerSection, listSection),
		slack.MsgOptionAsUser(true),
		slack.MsgOptionTS(event.TimeStamp),
	)

	if err != nil {
		return err
	}

	return nil
}
