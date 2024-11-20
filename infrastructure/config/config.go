package config

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/caarlos0/env/v6"
	"github.com/google/go-github/v50/github"
	"github.com/opensearch-project/opensearch-go/v2"
)

type Config struct {
	Debug         bool   `env:"DEBUG" envDefault:"false"`
	OpensearchUrl string `env:"OPEN_SEARCH_URL"`
	ModelName     string `env:"MODEL_NAME"`
	IndexName     string `env:"INDEX_NAME"`

	GitHubAppId          int64  `env:"GITHUB_APP_ID"`
	GithubInstallationId int64  `env:"GITHUB_INSTALLATION_ID"`
	GITHUB_PRIVATE_KEY   string `env:"GITHUB_PRIVATE_KEY"`
	GithubWebhookSecret  string `env:"GITHUB_WEBHOOK_SECRET"`

	JiraToken string `env:"JIRA_TOKEN"`
	JiraEmail string `env:"JIRA_EMAIL"`

	SlackSigningSecret string `env:"SLACK_SIGNING_SECRET"`
	SlackBotToken      string `env:"SLACK_BOT_TOKEN"`

	ModelId          string
	OpensearchClient *opensearch.Client
	GithubClient     *github.Client
}

func NewFromEnv(ctx context.Context) (Config, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}

	modelId, found := findModel(cfg, ctx)
	if found {
		cfg.ModelId = modelId
	}

	cfg.OpensearchClient, _ = opensearch.NewClient(opensearch.Config{
		Addresses: []string{cfg.OpensearchUrl},
	})

	decodedPrivateKey, err := base64.StdEncoding.DecodeString(cfg.GITHUB_PRIVATE_KEY)

	if err != nil {
		return cfg, err
	}

	itr, err := ghinstallation.New(http.DefaultTransport, cfg.GitHubAppId, cfg.GithubInstallationId, decodedPrivateKey)

	if err != nil {
		log.Printf("Error creating GitHub client: %v, using default client", err)
		cfg.GithubClient = github.NewClient(nil)
	} else {
		cfg.GithubClient = github.NewClient(&http.Client{Transport: itr})
	}

	return cfg, nil
}

func findModel(config Config, ctx context.Context) (string, bool) {
	var searchForModelRequest = []byte(`{
	  "query": {
		"term": {
		  "name.keyword": {
			"value": "` + config.ModelName + `"
		  }
		}
	  }
	}`)

	request, _ := http.NewRequestWithContext(ctx, "POST", config.OpensearchUrl+"/_plugins/_ml/models/_search", bytes.NewBuffer(searchForModelRequest))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	body, statusCode := doRequest(request)

	if statusCode != 200 {
		return "", false
	}
	var searchResult ModelSearchResponse
	if err := json.Unmarshal(body, &searchResult); err != nil {
		panic(err)
	}
	if searchResult.Hits.Total.Value == 0 {
		return "", false
	}

	return searchResult.Hits.Hits[0].Source.ModelId, true
}

func doRequest(request *http.Request) ([]byte, int) {
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return body, response.StatusCode
}

type ModelSearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index       string  `json:"_index"`
			Id          string  `json:"_id"`
			Version     int     `json:"_version"`
			SeqNo       int     `json:"_seq_no"`
			PrimaryTerm int     `json:"_primary_term"`
			Score       float64 `json:"_score"`
			Source      struct {
				ModelVersion    string `json:"model_version"`
				CreatedTime     int64  `json:"created_time"`
				ChunkNumber     int    `json:"chunk_number"`
				LastUpdatedTime int64  `json:"last_updated_time"`
				ModelFormat     string `json:"model_format"`
				Name            string `json:"name"`
				ModelId         string `json:"model_id"`
				TotalChunks     int    `json:"total_chunks"`
				Algorithm       string `json:"algorithm"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
