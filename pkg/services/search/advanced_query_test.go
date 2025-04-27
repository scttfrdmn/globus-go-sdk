// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"encoding/json"
	"testing"
)

func TestSimpleQuery(t *testing.T) {
	query := NewSimpleQuery("test query")

	// Check query type
	if query.Type() != QueryTypeSimple {
		t.Errorf("Expected query type %s, got %s", QueryTypeSimple, query.Type())
	}

	// Check JSON representation
	jsonMap := query.ToJSON()
	if len(jsonMap) != 1 {
		t.Errorf("Expected 1 key in JSON map, got %d", len(jsonMap))
	}

	queryString, ok := jsonMap["query_string"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected query_string to be a map, got %T", jsonMap["query_string"])
	}

	if queryString["query"] != "test query" {
		t.Errorf("Expected query = 'test query', got %v", queryString["query"])
	}
}

func TestMatchQuery(t *testing.T) {
	query := NewMatchQuery("title", "test")

	// Check query type
	if query.Type() != QueryTypeMatch {
		t.Errorf("Expected query type %s, got %s", QueryTypeMatch, query.Type())
	}

	// Check JSON representation
	jsonMap := query.ToJSON()
	if len(jsonMap) != 1 {
		t.Errorf("Expected 1 key in JSON map, got %d", len(jsonMap))
	}

	match, ok := jsonMap["match"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected match to be a map, got %T", jsonMap["match"])
	}

	if match["title"] != "test" {
		t.Errorf("Expected title = 'test', got %v", match["title"])
	}
}

func TestTermQuery(t *testing.T) {
	query := NewTermQuery("status", "active")

	// Check query type
	if query.Type() != QueryTypeTerm {
		t.Errorf("Expected query type %s, got %s", QueryTypeTerm, query.Type())
	}

	// Check JSON representation
	jsonMap := query.ToJSON()
	if len(jsonMap) != 1 {
		t.Errorf("Expected 1 key in JSON map, got %d", len(jsonMap))
	}

	term, ok := jsonMap["term"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected term to be a map, got %T", jsonMap["term"])
	}

	if term["status"] != "active" {
		t.Errorf("Expected status = 'active', got %v", term["status"])
	}
}

func TestRangeQuery(t *testing.T) {
	query := NewRangeQuery("created_at").
		WithGTE("2023-01-01").
		WithLT("2023-12-31").
		WithFormat("yyyy-MM-dd")

	// Check query type
	if query.Type() != QueryTypeRange {
		t.Errorf("Expected query type %s, got %s", QueryTypeRange, query.Type())
	}

	// Check JSON representation
	jsonMap := query.ToJSON()
	if len(jsonMap) != 1 {
		t.Errorf("Expected 1 key in JSON map, got %d", len(jsonMap))
	}

	rangeVal, ok := jsonMap["range"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected range to be a map, got %T", jsonMap["range"])
	}

	createdAt, ok := rangeVal["created_at"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected created_at to be a map, got %T", rangeVal["created_at"])
	}

	if createdAt["gte"] != "2023-01-01" {
		t.Errorf("Expected gte = '2023-01-01', got %v", createdAt["gte"])
	}

	if createdAt["lt"] != "2023-12-31" {
		t.Errorf("Expected lt = '2023-12-31', got %v", createdAt["lt"])
	}

	if createdAt["format"] != "yyyy-MM-dd" {
		t.Errorf("Expected format = 'yyyy-MM-dd', got %v", createdAt["format"])
	}
}

func TestBoolQuery(t *testing.T) {
	query := NewBoolQuery().
		AddMust(NewMatchQuery("title", "test")).
		AddMustNot(NewTermQuery("status", "deleted")).
		AddShould(NewTermQuery("tags", "important")).
		AddShould(NewTermQuery("tags", "urgent")).
		SetMinimumShouldMatch(1)

	// Check query type
	if query.Type() != QueryTypeBool {
		t.Errorf("Expected query type %s, got %s", QueryTypeBool, query.Type())
	}

	// Check JSON representation
	jsonMap := query.ToJSON()
	if len(jsonMap) != 1 {
		t.Errorf("Expected 1 key in JSON map, got %d", len(jsonMap))
	}

	boolVal, ok := jsonMap["bool"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected bool to be a map, got %T", jsonMap["bool"])
	}

	// Check must clause
	must, ok := boolVal["must"].([]interface{})
	if !ok {
		t.Fatalf("Expected must to be an array, got %T", boolVal["must"])
	}
	if len(must) != 1 {
		t.Errorf("Expected 1 must clause, got %d", len(must))
	}

	// Check must_not clause
	mustNot, ok := boolVal["must_not"].([]interface{})
	if !ok {
		t.Fatalf("Expected must_not to be an array, got %T", boolVal["must_not"])
	}
	if len(mustNot) != 1 {
		t.Errorf("Expected 1 must_not clause, got %d", len(mustNot))
	}

	// Check should clause
	should, ok := boolVal["should"].([]interface{})
	if !ok {
		t.Fatalf("Expected should to be an array, got %T", boolVal["should"])
	}
	if len(should) != 2 {
		t.Errorf("Expected 2 should clauses, got %d", len(should))
	}

	// Check minimum_should_match
	if boolVal["minimum_should_match"] != 1 {
		t.Errorf("Expected minimum_should_match = 1, got %v", boolVal["minimum_should_match"])
	}
}

func TestExistsQuery(t *testing.T) {
	query := NewExistsQuery("attachment")

	// Check query type
	if query.Type() != QueryTypeExists {
		t.Errorf("Expected query type %s, got %s", QueryTypeExists, query.Type())
	}

	// Check JSON representation
	jsonMap := query.ToJSON()
	if len(jsonMap) != 1 {
		t.Errorf("Expected 1 key in JSON map, got %d", len(jsonMap))
	}

	exists, ok := jsonMap["exists"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected exists to be a map, got %T", jsonMap["exists"])
	}

	if exists["field"] != "attachment" {
		t.Errorf("Expected field = 'attachment', got %v", exists["field"])
	}
}

func TestStructuredSearchRequest(t *testing.T) {
	// Create a structured search request
	request := &StructuredSearchRequest{
		IndexID: "test-index",
		Query:   NewMatchQuery("title", "test"),
		Options: &SearchOptions{
			Limit: 10,
			Sort:  []string{"created_at:desc"},
		},
		Extra: map[string]interface{}{
			"highlight": map[string]interface{}{
				"fields": map[string]interface{}{
					"title": map[string]interface{}{},
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Unmarshal to map to check fields
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check fields
	if result["index_id"] != "test-index" {
		t.Errorf("Expected index_id = 'test-index', got %v", result["index_id"])
	}

	if result["limit"] != float64(10) {
		t.Errorf("Expected limit = 10, got %v", result["limit"])
	}

	match, ok := result["match"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected match to be a map, got %T", result["match"])
	}

	if match["title"] != "test" {
		t.Errorf("Expected title = 'test', got %v", match["title"])
	}

	highlight, ok := result["highlight"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected highlight to be a map, got %T", result["highlight"])
	}

	fields, ok := highlight["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected fields to be a map, got %T", highlight["fields"])
	}

	if _, ok := fields["title"].(map[string]interface{}); !ok {
		t.Fatalf("Expected title to be a map, got %T", fields["title"])
	}
}

func TestQueryParser(t *testing.T) {
	parser := &QueryParser{}

	// Test simple query
	query, err := parser.ParseQuery("test query")
	if err != nil {
		t.Fatalf("Failed to parse simple query: %v", err)
	}
	if _, ok := query.(*SimpleQuery); !ok {
		t.Errorf("Expected SimpleQuery, got %T", query)
	}

	// Test term query
	query, err = parser.ParseQuery("status:active")
	if err != nil {
		t.Fatalf("Failed to parse term query: %v", err)
	}
	if termQuery, ok := query.(*TermQuery); ok {
		if termQuery.Field != "status" {
			t.Errorf("Expected field = 'status', got %s", termQuery.Field)
		}
		if termQuery.Value != "active" {
			t.Errorf("Expected value = 'active', got %v", termQuery.Value)
		}
	} else {
		t.Errorf("Expected TermQuery, got %T", query)
	}

	// Test wildcard query
	query, err = parser.ParseQuery("title:test*")
	if err != nil {
		t.Fatalf("Failed to parse wildcard query: %v", err)
	}
	if wildcardQuery, ok := query.(*WildcardQuery); ok {
		if wildcardQuery.Field != "title" {
			t.Errorf("Expected field = 'title', got %s", wildcardQuery.Field)
		}
		if wildcardQuery.Wildcard != "test*" {
			t.Errorf("Expected wildcard = 'test*', got %s", wildcardQuery.Wildcard)
		}
	} else {
		t.Errorf("Expected WildcardQuery, got %T", query)
	}

	// Test range query
	query, err = parser.ParseQuery("created_at:[2023-01-01 TO 2023-12-31]")
	if err != nil {
		t.Fatalf("Failed to parse range query: %v", err)
	}
	if rangeQuery, ok := query.(*RangeQuery); ok {
		if rangeQuery.Field != "created_at" {
			t.Errorf("Expected field = 'created_at', got %s", rangeQuery.Field)
		}
		if rangeQuery.GTE != "2023-01-01" {
			t.Errorf("Expected GTE = '2023-01-01', got %v", rangeQuery.GTE)
		}
		if rangeQuery.LTE != "2023-12-31" {
			t.Errorf("Expected LTE = '2023-12-31', got %v", rangeQuery.LTE)
		}
	} else {
		t.Errorf("Expected RangeQuery, got %T", query)
	}

	// Test open-ended range query
	query, err = parser.ParseQuery("created_at:[2023-01-01 TO *]")
	if err != nil {
		t.Fatalf("Failed to parse open-ended range query: %v", err)
	}
	if rangeQuery, ok := query.(*RangeQuery); ok {
		if rangeQuery.Field != "created_at" {
			t.Errorf("Expected field = 'created_at', got %s", rangeQuery.Field)
		}
		if rangeQuery.GTE != "2023-01-01" {
			t.Errorf("Expected GTE = '2023-01-01', got %v", rangeQuery.GTE)
		}
		if rangeQuery.LTE != nil {
			t.Errorf("Expected LTE = nil, got %v", rangeQuery.LTE)
		}
	} else {
		t.Errorf("Expected RangeQuery, got %T", query)
	}
}
