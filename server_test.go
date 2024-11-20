//go:build integration

package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/shopwarelabs/jira-issue-bot/domain/search"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"

	"github.com/steinfletcher/apitest"
)

// Not possible to split into multiple tests, as by default go executes the tests in parallel
func TestServerIntegration(t *testing.T) {
	ctx := context.Background()
	cfg, _ := config.NewFromEnv(ctx)

	createServer(cfg, ctx)

	// test UI serving
	apitest.New().Handler(http.DefaultServeMux).
		Get("/app/").
		Expect(t).
		Status(http.StatusOK).
		BodyFromFile("ui/dist/index.html").
		End()

	// index one jira issue
	apitest.New().Handler(http.DefaultServeMux).
		Post("/webhook/jira").
		Body(`{
			"webhookEvent": "jira:issue_updated",
			"issue": {	
				"key": "NEXT-1",		
				"fields": {	
					"summary": "Test Issue",	
					"description": "This is a test issue",	
					"issuetype": {	
						"name": "Bug"	
					},	
					"project": {	
						"key": "NEXT"	
					},
					"status": {					
						"name": "Open"	
					},
					"reporter": {
						"displayName": "Author"
					}
				}
			}
		}`).
		Expect(t).
		Status(http.StatusOK).
		End()

	// index one github issue
	apitest.New().Handler(http.DefaultServeMux).
		Post("/webhook/github").
		Header("X-GitHub-Event", "issues").
		Body(`{
			"action": "edited",
			"issue": {
				"number": 1,
				"title": "Github Issues are fancy",
				"body": "Fancy that github supports webhook issues",
				"state": "open",
				"labels": [	
					{
						"name": "bug"	
					}
				],	
				"user": {	
					"login": "author"
				},
				"created_at": "2021-01-01T00:00:00Z"
			}
		}`).
		Expect(t).
		Status(http.StatusOK).
		End()

	// wait for indexing to finish
	time.Sleep(1 * time.Second)

	// search per API
	searchResult := apitest.New().Handler(http.DefaultServeMux).
		Post("/api/search").
		Query("title", "Fancy github webhook").
		Expect(t).
		Status(http.StatusOK).
		End()

	body, err := io.ReadAll(searchResult.Response.Body)
	defer searchResult.Response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	var result search.SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}

	if result.Hits.Total.Value != 1 {
		t.Fatalf("expected 1 hit, got %d", result.Hits.Total.Value)
	}

	if result.Hits.Hits[0].ID != "GH-1" {
		t.Fatalf("expected GH-1, got %s", result.Hits.Hits[0].ID)
	}

	if result.Hits.Hits[0].Source.Title != "Github Issues are fancy" {
		t.Fatalf("expected Github Issues are fancy, got %s", result.Hits.Hits[0].Source.Title)
	}

	// index duplicate to first jira issue
	apitest.New().Handler(http.DefaultServeMux).
		Post("/webhook/jira").
		Body(`{
			"webhookEvent": "jira:issue_updated",
			"issue": {	
				"key": "NEXT-2",		
				"fields": {	
					"summary": "Test Issue",	
					"description": "This is a test issue",	
					"issuetype": {	
						"name": "Bug"	
					},	
					"project": {	
						"key": "NEXT"	
					},
					"status": {					
						"name": "Open"	
					},
					"reporter": {
						"displayName": "Author"
					}
				}
			}
		}`).
		Expect(t).
		Status(http.StatusOK).
		End()

	// wait for indexing to finish
	time.Sleep(1 * time.Second)

	// search per ID
	searchResult = apitest.New().Handler(http.DefaultServeMux).
		Post("/api/search-id").
		Query("id", "NEXT-1").
		Expect(t).
		Status(http.StatusOK).
		End()

	body, err = io.ReadAll(searchResult.Response.Body)
	defer searchResult.Response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}

	if result.Hits.Total.Value != 1 {
		t.Fatalf("expected 1 hit, got %d", result.Hits.Total.Value)
	}

	if result.Hits.Hits[0].ID != "NEXT-2" {
		t.Fatalf("expected NEXT-2, got %s", result.Hits.Hits[0].ID)
	}

	if result.Hits.Hits[0].Source.Title != "Test Issue" {
		t.Fatalf("expected Test Issue, got %s", result.Hits.Hits[0].Source.Title)
	}
}
