package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/shopwarelabs/jira-issue-bot/domain/github_connector"
	"github.com/shopwarelabs/jira-issue-bot/domain/slack_connector"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/cmd"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"
	"github.com/shopwarelabs/jira-issue-bot/infrastructure/logging"

	"github.com/MadAppGang/httplog"
	"github.com/google/go-github/v50/github"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	RunE: func(command *cobra.Command, args []string) error {
		cfg := command.Context().Value(cmd.ConfigKey{}).(config.Config)

		server := createServer(cfg, command.Context())

		return server.ListenAndServe()
	},
}

func createServer(cfg config.Config, context context.Context) *http.Server {
	logger := logging.FromContext(context)
	loggerWithFormatter := httplog.LoggerWithFormatter(httplog.DefaultLogFormatter)

	http.Handle("/webhook/github", loggerWithFormatter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload []byte
		var err error

		if os.Getenv("CI") == "true" {
			payload, err = io.ReadAll(r.Body)
		} else {
			payload, err = github.ValidatePayload(r, []byte(cfg.GithubWebhookSecret))
		}

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		event, err := github.ParseWebHook(github.WebHookType(r), payload)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch event := event.(type) {
		case *github.IssuesEvent:
			if err = github_connector.HandleGithubIssueEvent(event, cfg, logger); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case *github.PullRequestEvent:
			if err = github_connector.HandleGithubPREvent(event, cfg, logger); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	})))

	http.Handle("/slack/command", loggerWithFormatter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, cfg.SlackSigningSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
		command, err := slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch command.Command {
		case "/issues":
		case "/aiaiai":
			message, err := slack_connector.OnIssuesCommand(command, cfg, logger)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			JSONResp(w, message, logger)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})))

	http.Handle("/slack/event", loggerWithFormatter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		verifier, err := slack.NewSecretsVerifier(r.Header, cfg.SlackSigningSecret)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, err := verifier.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal(body, &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			_, _ = w.Write([]byte(r.Challenge))
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			mentionEvent, ok := innerEvent.Data.(*slackevents.AppMentionEvent)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err := slack_connector.OnMention(mentionEvent, cfg, logger)

			if err != nil {
				_, _ = w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	})))

	return &http.Server{
		Addr:              ":8000",
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func JSONError(w http.ResponseWriter, err error, code int, logger *zap.SugaredLogger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	err = json.NewEncoder(w).Encode(struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   true,
		Message: err.Error(),
	})

	if err != nil {
		logger.Errorf("error encoding json error: %s", err)
	}
}

func JSONResp(w http.ResponseWriter, resp any, logger *zap.SugaredLogger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(resp)

	if err != nil {
		logger.Errorf("error encoding json: %s", err)
	}
}
