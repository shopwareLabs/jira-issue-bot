package search

import (
	"strings"
	"testing"
)

func TestParseFilterWithEmptyFilter(t *testing.T) {
	parsedQuery := parseFilter("title", "description", "modelId", SearchFilter{})
	parsedQueryString := parsedQuery.String()

	expected := `{
	"min_score": 1.8,
    "query": {
        "bool": {
            "should": [
				{
					"script_score": {
						"query": {
							"neural": {
								"title_embedding": { "model_id": "modelId", "k": 100, "query_text": title }
							}
						},
						"script": { "source": "_score * 1.8"}
					}
				},
				{
					"script_score": {
						"query": {
							"neural": {
								"description_embedding": { "model_id": "modelId", "k": 100, "query_text": description }
							}
						},
						"script": { "source": "_score * 1.5" }
					}
				}
            ]
        }
    }
}`

	assertQuery(t, expected, parsedQueryString)
}

func TestParseFilterWithAllFilterOptions(t *testing.T) {
	parsedQuery := parseFilter("title", "description", "modelId", SearchFilter{
		ExcludedDocumentId: "excludedDocumentId",
		OnlyPublic:         true,
		Source:             "source",
	})
	parsedQueryString := parsedQuery.String()

	expected := `{
	"min_score": 1.8,
    "query": {
        "bool": {
            "must_not": [
				{ "ids": { "values": ["excludedDocumentId"] }}
            ],
            "must": [
            	{ "match": { "public": true }},
				{ "match": { "source": "source" }}
            ],
            "should": [
				{
					"script_score": {
						"query": {
							"neural": {
								"title_embedding": { "model_id": "modelId", "k": 100, "query_text": title }
							}
						},
						"script": { "source": "_score * 1.8"}
					}
				},
				{
					"script_score": {
						"query": {
							"neural": {
								"description_embedding": { "model_id": "modelId", "k": 100, "query_text": description }
							}
						},
						"script": { "source": "_score * 1.5" }
					}
				}
            ]
        }
    }
}`

	assertQuery(t, expected, parsedQueryString)
}

func TestParseFilterWithExcludedDocumentId(t *testing.T) {
	parsedQuery := parseFilter("title", "description", "modelId", SearchFilter{
		ExcludedDocumentId: "excludedDocumentId",
	})
	parsedQueryString := parsedQuery.String()

	expected := `{
	"min_score": 1.8,
    "query": {
        "bool": {
            "must_not": [
				{ "ids": { "values": ["excludedDocumentId"] }}
            ],
            "should": [
				{
					"script_score": {
						"query": {
							"neural": {
								"title_embedding": { "model_id": "modelId", "k": 100, "query_text": title }
							}
						},
						"script": { "source": "_score * 1.8"}
					}
				},
				{
					"script_score": {
						"query": {
							"neural": {
								"description_embedding": { "model_id": "modelId", "k": 100, "query_text": description }
							}
						},
						"script": { "source": "_score * 1.5" }
					}
				}
            ]
        }
    }
}`

	assertQuery(t, expected, parsedQueryString)
}

func TestParseFilterWithSource(t *testing.T) {
	parsedQuery := parseFilter("title", "description", "modelId", SearchFilter{
		Source: "source",
	})
	parsedQueryString := parsedQuery.String()

	expected := `{
	"min_score": 1.8,
    "query": {
        "bool": {
            "must": [
				{ "match": { "source": "source" }}
            ],
            "should": [
				{
					"script_score": {
						"query": {
							"neural": {
								"title_embedding": { "model_id": "modelId", "k": 100, "query_text": title }
							}
						},
						"script": { "source": "_score * 1.8"}
					}
				},
				{
					"script_score": {
						"query": {
							"neural": {
								"description_embedding": { "model_id": "modelId", "k": 100, "query_text": description }
							}
						},
						"script": { "source": "_score * 1.5" }
					}
				}
            ]
        }
    }
}`

	assertQuery(t, expected, parsedQueryString)
}

func TestParseFilterWithOnlyPublicFilter(t *testing.T) {
	parsedQuery := parseFilter("title", "description", "modelId", SearchFilter{
		OnlyPublic: true,
	})
	parsedQueryString := parsedQuery.String()

	expected := `{
	"min_score": 1.8,
    "query": {
        "bool": {
            "must": [
            	{ "match": { "public": true }}
            ],
            "should": [
				{
					"script_score": {
						"query": {
							"neural": {
								"title_embedding": { "model_id": "modelId", "k": 100, "query_text": title }
							}
						},
						"script": { "source": "_score * 1.8"}
					}
				},
				{
					"script_score": {
						"query": {
							"neural": {
								"description_embedding": { "model_id": "modelId", "k": 100, "query_text": description }
							}
						},
						"script": { "source": "_score * 1.5" }
					}
				}
            ]
        }
    }
}`

	assertQuery(t, expected, parsedQueryString)
}

func assertQuery(t *testing.T, expected string, actual string) {
	t.Helper()

	if removeWhitespace(expected) != removeWhitespace(actual) {
		t.Errorf("expected query string %s, got %s", expected, actual)
	}
}

// removes whitespace from string, so that changes to the formatting don't break the tests.
func removeWhitespace(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")

	return s
}
