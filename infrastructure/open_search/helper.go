package open_search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopwarelabs/jira-issue-bot/infrastructure/config"

	"go.uber.org/zap"
)

func LoadModel(config config.Config, ctx context.Context, logger *zap.SugaredLogger) error {
	request, _ := http.NewRequestWithContext(ctx, "POST", config.OpensearchUrl+"/_plugins/_ml/models/"+config.ModelId+"/_load", nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	body := doRequest(request)

	var result ModelLoadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	for result.Status != "CREATED" {
		if result.Status == "FAILED" {
			panic("Model load failed")
		}

		logger.Debug("Waiting for model to be loaded. Current state: ", result.Status)
		time.Sleep(2 * time.Second)

		body = doRequest(request)

		if err := json.Unmarshal(body, &result); err != nil {
			panic(err)
		}
	}

	logger.Debugf("Model \"%s\" loaded", config.ModelId)

	return nil
}

func CreatePipeline(config config.Config, ctx context.Context, logger *zap.SugaredLogger) {
	var jsonData = []byte(`{
	  "description": "Jira NLP pipeline",
	  "processors" : [
		{
		  "text_embedding": {
			"model_id": "` + config.ModelId + `",
			"field_map": {
			   "title": "title_embedding",
			   "description": "description_embedding"
			}
		  }
		}
	  ]
	}`)

	request, _ := http.NewRequestWithContext(ctx, "PUT", config.OpensearchUrl+"/_ingest/pipeline/nlp-pipeline", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	_ = doRequest(request)

	logger.Debug("Pipeline created")
}

func CreateIndex(config config.Config, ctx context.Context, logger *zap.SugaredLogger) {
	var jsonData = []byte(`{
		"settings": {
		"index.knn": true,
		"default_pipeline": "nlp-pipeline"
	  },
	  "mappings": {
		"_source": {
		  "excludes": [
			"title_embedding",
			"description_embedding"
		  ]
		},
		"properties": {
		  "title_embedding": {
			"type": "knn_vector",
			"dimension": 384,
			"method": {
			  "name": "hnsw"
			}
		  },
		  "title": {
			"type": "text"
		  },
		  "description_embedding": {
			"type": "knn_vector",
			"dimension": 384,
			"method": {
			  "name": "hnsw"
			}
		  },
		  "description": {
			"type": "text"
		  }
		}
	  }
	}`)

	request, _ := http.NewRequestWithContext(ctx, "PUT", config.OpensearchUrl+"/"+config.IndexName, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	_ = doRequest(request)

	logger.Debugf("Index \"%s\" created", config.IndexName)
}

func CreateModel(config config.Config, ctx context.Context, logger *zap.SugaredLogger) string {
	var jsonData = []byte(`{
			 "name": "` + config.ModelName + `",
  			 "version": "1.0.0",
             "description": "test model",
             "model_format": "TORCH_SCRIPT",
             "model_config": {
    			"model_type": "bert",
    			"embedding_dimension": 384,
    			"framework_type": "sentence_transformers"
  			},
			"model_content_hash_value": "c15f0d2e62d872be5b5bc6c84d2e0f4921541e29fefbef51d59cc10a8ae30e0f",
  			"url": "https://artifacts.opensearch.org/models/ml-models/huggingface/sentence-transformers/all-MiniLM-L6-v2/1.0.1/torch_script/sentence-transformers_all-MiniLM-L6-v2-1.0.1-torch_script.zip"
		}`)

	request, _ := http.NewRequestWithContext(ctx, "POST", config.OpensearchUrl+"/_plugins/_ml/models/_upload", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	body := doRequest(request)

	var result UploadModelResponse
	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	return getModelId(result.TaskId, config, ctx, logger)
}

func getModelId(taskId string, config config.Config, ctx context.Context, logger *zap.SugaredLogger) string {
	var url = config.OpensearchUrl + "/_plugins/_ml/tasks/" + taskId
	request, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	body := doRequest(request)

	var result TaskResponse
	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	for result.State != "COMPLETED" {
		if result.State == "FAILED" {
			fmt.Println(string(body))
			panic("Model upload failed")
		}

		logger.Debug("Waiting for model to be uploaded. Current state: ", result.State)
		time.Sleep(2 * time.Second)

		body = doRequest(request)

		if err := json.Unmarshal(body, &result); err != nil {
			panic(err)
		}
	}

	return result.ModelId
}

func doRequest(request *http.Request) []byte {
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return body
}

type UploadModelResponse struct {
	TaskId string `json:"task_id"`
}

type TaskResponse struct {
	ModelId string `json:"model_id"`
	State   string `json:"state"`
}

type ModelLoadResponse struct {
	TaskId string `json:"task_id"`
	Status string `json:"status"`
}
