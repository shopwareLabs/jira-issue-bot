package slack_connector

import (
	"fmt"
	"strings"

	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/slack-go/slack"
)

func noResultsSectionBlock() *slack.SectionBlock {
	noIssuesText := slack.NewTextBlockObject(
		"mrkdwn",
		"✅ We didn't find any existing issues that are related to your topic",
		false,
		false,
	)
	return slack.NewSectionBlock(noIssuesText, nil, nil)
}

func headerSectionBlock() *slack.SectionBlock {
	headerText := slack.NewTextBlockObject(
		"mrkdwn",
		"🤖 We found the following existing issues which may help or are related to your topic: ",
		false,
		false,
	)
	return slack.NewSectionBlock(headerText, nil, nil)
}

func resultListSectionBlock(result *search.SearchResponse) *slack.SectionBlock {
	output := strings.Builder{}
	for _, hit := range result.Hits.Hits {
		switch hit.Source.Source {
		case "jira":
			output.WriteString(fmt.Sprintf("• <%s|%s>\n", "https://shopware.atlassian.net/browse/"+hit.ID, hit.Source.Title))
		case "github":
			link := "https://github.com/shopware/platform/issues/" + strings.Replace(hit.ID, "GH-", "", 1)
			output.WriteString(fmt.Sprintf("• <%s|%s>\n", link, hit.Source.Title))
		}
	}

	listText := slack.NewTextBlockObject("mrkdwn", output.String(), false, false)
	listSection := slack.NewSectionBlock(listText, nil, nil)
	return listSection
}
