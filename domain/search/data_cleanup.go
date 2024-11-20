package search

import (
	"regexp"
	"strings"

	stripmd "github.com/writeas/go-strip-markdown"
)

var (
	// Issue tracker template.
	description      = regexp.MustCompile(`(?m)Description:\s*\n*`)
	environment      = regexp.MustCompile(`(?m)Environment:\s*\n*`)
	stepsToReproduce = regexp.MustCompile(`(?m)Steps to reproduce:\s*\n*`)
	expectedResult   = regexp.MustCompile(`(?m)Expected result:\s*\n*`)
	actualResult     = regexp.MustCompile(`(?m)Actual result:\s*\n*`)

	// Github issue template.
	github          = regexp.MustCompile(`(?s)### PHP Version.*### Expected behaviour`)
	actualBehaviour = regexp.MustCompile(`(?m)### Actual behaviour\s*\n*`)
	reproducer      = regexp.MustCompile(`(?m)### How to reproduce\s*\n*`)

	// Github PR template.
	githubPR          = regexp.MustCompile(`(?s)<!--.*-->`)
	githubOrder       = regexp.MustCompile(`(?m)\s[1a]\.\s*\n*$`)
	githubWhy         = regexp.MustCompile(`(?m)1\. Why is this change necessary\?`)
	githubWhat        = regexp.MustCompile(`(?m)2\. What does this change do, exactly\?`)
	githubDescription = regexp.MustCompile(`(?m)3\. Describe each step to reproduce the issue or behaviour\.`)
	githubLinks       = regexp.MustCompile(`(?m)4\. Please link to the relevant issues \(if any\)\.`)
	githubChecklist   = regexp.MustCompile(`(?m)5\. Checklist`)
	githubChecks      = regexp.MustCompile(`(?m)- \[[ x]].*`)
)

// CleanupString removes markdown and trims the string.
func CleanupString(text string) string {
	text = description.ReplaceAllString(text, "")
	text = environment.ReplaceAllString(text, "")
	text = stepsToReproduce.ReplaceAllString(text, "")
	text = expectedResult.ReplaceAllString(text, "")
	text = actualResult.ReplaceAllString(text, "")

	text = github.ReplaceAllString(text, "")
	text = actualBehaviour.ReplaceAllString(text, "")
	text = reproducer.ReplaceAllString(text, "")

	text = githubPR.ReplaceAllString(text, "")
	text = githubOrder.ReplaceAllString(text, "")
	text = githubWhy.ReplaceAllString(text, "")
	text = githubWhat.ReplaceAllString(text, "")
	text = githubDescription.ReplaceAllString(text, "")
	text = githubLinks.ReplaceAllString(text, "")
	text = githubChecklist.ReplaceAllString(text, "")
	text = githubChecks.ReplaceAllString(text, "")

	text = stripmd.Strip(text)
	text = strings.Trim(text, " \n\t\r")

	return text
}
