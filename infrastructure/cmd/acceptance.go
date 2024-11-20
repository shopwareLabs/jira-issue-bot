package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/samber/lo"
	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/spf13/cobra"
)

var acceptanceCommand = &cobra.Command{
	Use:   "test",
	Short: "Run the acceptance test suite to see how good our search is performing",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cmd.Context().Value(ConfigKey{}).(config.Config)

		files, err := os.ReadDir("duplicates")
		if err != nil {
			return fmt.Errorf("failed reading directory: %w", err)
		}

		var wg sync.WaitGroup

		for _, file := range files {
			var issue Issue

			readFile, _ := os.ReadFile("duplicates/" + file.Name())
			if err := json.Unmarshal(readFile, &issue); err != nil {
				log.Println(err)
				os.Exit(1)
			}
			issue.GithubIssue = "GH-" + file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]

			wg.Add(1)

			go func(issue Issue) {
				var matches []string

				result, err := search.Search(
					issue.Title,
					issue.Description,
					search.SearchFilter{ExcludedDocumentId: issue.GithubIssue},
					cfg,
				)
				if err != nil {
					log.Println("search failed", err)
					os.Exit(1)
				}

				for _, hit := range result.Hits.Hits {
					matches = append(matches, hit.ID)
				}

				expected, actual := lo.Difference(issue.Matches, matches)

				if len(expected) > 0 {
					log.Printf("Expected that jira issues \"%s\" to be found for issue \"%s\"\n", strings.Join(expected, ", "), issue.GithubIssue)
				}
				if len(actual) > 0 {
					log.Printf("Unexpected jira issues \"%s\" found for issue \"%s\"\n", strings.Join(actual, ", "), issue.GithubIssue)
				}

				defer wg.Done()
			}(issue)
		}

		wg.Wait()

		return nil
	},
}

func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(acceptanceCommand)
	rootCmd.AddCommand(dryRunCommand)
	rootCmd.AddCommand(initOpensearchCommand)
	rootCmd.AddCommand(loadModelCommand)
	rootCmd.AddCommand(createIndexCommand)
}

type Issue struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Matches     []string `json:"matches"`
	GithubIssue string   `json:"omitempty"`
}
