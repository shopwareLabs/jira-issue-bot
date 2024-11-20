package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"
	"github.com/spf13/cobra"
)

var dryRunCommand = &cobra.Command{
	Use:   "dry-run",
	Short: "Perform a dry-run on the existing issues and find duplicates for those.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cmd.Context().Value(ConfigKey{}).(config.Config)
		logger := logging.FromContext(cmd.Context())

		req := opensearchapi.SearchRequest{
			Index: []string{"issues"},
			Body: strings.NewReader(`{
				"size": 1000,
				"query": {
					"bool": {
						"must": [
							{										
								"term": {
									"status.keyword": "open"		
								}
							},
							{
								"term": {	
									"source.keyword": "github"	
								}
							}
						]
					}
				}
			}`),
		}

		resp, err := req.Do(context.Background(), cfg.OpensearchClient)
		if err != nil {
			return fmt.Errorf("failed to execute search: %w", err)
		}

		var result search.SearchResponse

		body, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to decode opensearch response: %w", err)
		}

		var wg sync.WaitGroup
		guard := make(chan struct{}, 10)

		for _, issue := range result.Hits.Hits {
			wg.Add(1)
			guard <- struct{}{}

			go func(issue search.IssueResult) {
				var matches []string

				result, err := search.Search(
					issue.Source.Title,
					issue.Source.Title,
					search.SearchFilter{ExcludedDocumentId: issue.ID},
					cfg)
				if err != nil {
					logger.Error("search failed", err)
					os.Exit(1)
				}

				for _, hit := range result.Hits.Hits {
					if hit.Score < 2.0 {
						continue
					}

					matches = append(matches, hit.ID+" (Score:"+fmt.Sprint(hit.Score)+")")
				}

				if len(matches) > 0 {
					logger.Error("Issue \"%s\" might be duplictated by: \"%s\"\n", issue.ID, strings.Join(matches, ", "))
				}

				<-guard
				defer wg.Done()
			}(issue)
		}

		wg.Wait()

		return nil
	},
}
