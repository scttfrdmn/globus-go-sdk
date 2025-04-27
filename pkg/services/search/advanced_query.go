// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"encoding/json"
	"fmt"
	"strings"
)

// QueryType defines the type of query
type QueryType string

// Query types
const (
	QueryTypeSimple   QueryType = "simple"
	QueryTypeMatch    QueryType = "match"
	QueryTypeTerm     QueryType = "term"
	QueryTypeRange    QueryType = "range"
	QueryTypeBool     QueryType = "bool"
	QueryTypePrefix   QueryType = "prefix"
	QueryTypeWildcard QueryType = "wildcard"
	QueryTypeExists   QueryType = "exists"
	QueryTypeGeo      QueryType = "geo_distance"
)

// RangeOperator defines the range operator type
type RangeOperator string

// Range operators
const (
	RangeGT  RangeOperator = "gt"
	RangeGTE RangeOperator = "gte"
	RangeLT  RangeOperator = "lt"
	RangeLTE RangeOperator = "lte"
)

// BoolOperator defines the boolean operator type
type BoolOperator string

// Boolean operators
const (
	BoolMust    BoolOperator = "must"
	BoolMustNot BoolOperator = "must_not"
	BoolShould  BoolOperator = "should"
)

// Query is the interface all query types must implement
type Query interface {
	Type() QueryType
	ToJSON() map[string]interface{}
}

// SimpleQuery is a simple string query
type SimpleQuery struct {
	Q string
}

// Type returns the query type
func (q *SimpleQuery) Type() QueryType {
	return QueryTypeSimple
}

// ToJSON converts the query to JSON
func (q *SimpleQuery) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"query_string": map[string]interface{}{
			"query": q.Q,
		},
	}
}

// MatchQuery is a match query
type MatchQuery struct {
	Field string
	Value interface{}
}

// Type returns the query type
func (q *MatchQuery) Type() QueryType {
	return QueryTypeMatch
}

// ToJSON converts the query to JSON
func (q *MatchQuery) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"match": map[string]interface{}{
			q.Field: q.Value,
		},
	}
}

// TermQuery is a term query
type TermQuery struct {
	Field string
	Value interface{}
}

// Type returns the query type
func (q *TermQuery) Type() QueryType {
	return QueryTypeTerm
}

// ToJSON converts the query to JSON
func (q *TermQuery) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"term": map[string]interface{}{
			q.Field: q.Value,
		},
	}
}

// RangeQuery is a range query
type RangeQuery struct {
	Field    string
	GT       interface{}
	GTE      interface{}
	LT       interface{}
	LTE      interface{}
	Format   string
	TimeZone string
	Relation string
}

// Type returns the query type
func (q *RangeQuery) Type() QueryType {
	return QueryTypeRange
}

// ToJSON converts the query to JSON
func (q *RangeQuery) ToJSON() map[string]interface{} {
	rangeValues := make(map[string]interface{})

	if q.GT != nil {
		rangeValues[string(RangeGT)] = q.GT
	}
	if q.GTE != nil {
		rangeValues[string(RangeGTE)] = q.GTE
	}
	if q.LT != nil {
		rangeValues[string(RangeLT)] = q.LT
	}
	if q.LTE != nil {
		rangeValues[string(RangeLTE)] = q.LTE
	}

	if q.Format != "" {
		rangeValues["format"] = q.Format
	}
	if q.TimeZone != "" {
		rangeValues["time_zone"] = q.TimeZone
	}
	if q.Relation != "" {
		rangeValues["relation"] = q.Relation
	}

	return map[string]interface{}{
		"range": map[string]interface{}{
			q.Field: rangeValues,
		},
	}
}

// BoolQuery is a boolean query
type BoolQuery struct {
	Must               []Query
	MustNot            []Query
	Should             []Query
	MinimumShouldMatch int
}

// Type returns the query type
func (q *BoolQuery) Type() QueryType {
	return QueryTypeBool
}

// ToJSON converts the query to JSON
func (q *BoolQuery) ToJSON() map[string]interface{} {
	boolValues := make(map[string]interface{})

	if len(q.Must) > 0 {
		mustQueries := make([]interface{}, len(q.Must))
		for i, query := range q.Must {
			mustQueries[i] = query.ToJSON()
		}
		boolValues["must"] = mustQueries
	}

	if len(q.MustNot) > 0 {
		mustNotQueries := make([]interface{}, len(q.MustNot))
		for i, query := range q.MustNot {
			mustNotQueries[i] = query.ToJSON()
		}
		boolValues["must_not"] = mustNotQueries
	}

	if len(q.Should) > 0 {
		shouldQueries := make([]interface{}, len(q.Should))
		for i, query := range q.Should {
			shouldQueries[i] = query.ToJSON()
		}
		boolValues["should"] = shouldQueries
	}

	if q.MinimumShouldMatch > 0 {
		boolValues["minimum_should_match"] = q.MinimumShouldMatch
	}

	return map[string]interface{}{
		"bool": boolValues,
	}
}

// ExistsQuery checks if a field exists
type ExistsQuery struct {
	Field string
}

// Type returns the query type
func (q *ExistsQuery) Type() QueryType {
	return QueryTypeExists
}

// ToJSON converts the query to JSON
func (q *ExistsQuery) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"exists": map[string]interface{}{
			"field": q.Field,
		},
	}
}

// PrefixQuery matches documents with terms starting with a prefix
type PrefixQuery struct {
	Field  string
	Prefix string
}

// Type returns the query type
func (q *PrefixQuery) Type() QueryType {
	return QueryTypePrefix
}

// ToJSON converts the query to JSON
func (q *PrefixQuery) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"prefix": map[string]interface{}{
			q.Field: q.Prefix,
		},
	}
}

// WildcardQuery matches documents with terms matching a wildcard pattern
type WildcardQuery struct {
	Field    string
	Wildcard string
}

// Type returns the query type
func (q *WildcardQuery) Type() QueryType {
	return QueryTypeWildcard
}

// ToJSON converts the query to JSON
func (q *WildcardQuery) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"wildcard": map[string]interface{}{
			q.Field: q.Wildcard,
		},
	}
}

// GeoDistanceQuery matches documents with geo points within a distance
type GeoDistanceQuery struct {
	Field    string
	Distance string
	Lat      float64
	Lon      float64
}

// Type returns the query type
func (q *GeoDistanceQuery) Type() QueryType {
	return QueryTypeGeo
}

// ToJSON converts the query to JSON
func (q *GeoDistanceQuery) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"geo_distance": map[string]interface{}{
			"distance": q.Distance,
			q.Field: map[string]interface{}{
				"lat": q.Lat,
				"lon": q.Lon,
			},
		},
	}
}

// StructuredSearchRequest represents a structured search query request
type StructuredSearchRequest struct {
	IndexID string                 `json:"index_id"`
	Query   Query                  `json:"-"`
	Options *SearchOptions         `json:"options,omitempty"`
	Extra   map[string]interface{} `json:"-"`
}

// MarshalJSON custom JSON marshalling for StructuredSearchRequest
func (r *StructuredSearchRequest) MarshalJSON() ([]byte, error) {
	if r.Query == nil {
		return nil, fmt.Errorf("query is required")
	}

	// Build the request map
	requestMap := map[string]interface{}{
		"index_id": r.IndexID,
	}

	// Add the query
	queryMap := r.Query.ToJSON()
	for k, v := range queryMap {
		requestMap[k] = v
	}

	// Add options
	if r.Options != nil {
		if r.Options.Limit > 0 {
			requestMap["limit"] = r.Options.Limit
		} else if r.Options.PageSize > 0 {
			requestMap["limit"] = r.Options.PageSize
		}

		if r.Options.Offset > 0 {
			requestMap["offset"] = r.Options.Offset
		}

		if r.Options.PageToken != "" {
			requestMap["page_token"] = r.Options.PageToken
		}

		if len(r.Options.Sort) > 0 {
			requestMap["sort"] = r.Options.Sort
		}

		if r.Options.Filter != "" {
			requestMap["filter"] = r.Options.Filter
		}

		if len(r.Options.Facets) > 0 {
			requestMap["facets"] = r.Options.Facets
		}

		if r.Options.FacetSize > 0 {
			requestMap["facet_size"] = r.Options.FacetSize
		}

		if r.Options.IncludeAllContent {
			requestMap["include_all_content"] = true
		}

		if r.Options.ByPath != "" {
			requestMap["by_path"] = r.Options.ByPath
		}
	}

	// Add any extra parameters
	for k, v := range r.Extra {
		requestMap[k] = v
	}

	return json.Marshal(requestMap)
}

// NewSimpleQuery creates a new simple query
func NewSimpleQuery(queryString string) *SimpleQuery {
	return &SimpleQuery{
		Q: queryString,
	}
}

// NewMatchQuery creates a new match query
func NewMatchQuery(field string, value interface{}) *MatchQuery {
	return &MatchQuery{
		Field: field,
		Value: value,
	}
}

// NewTermQuery creates a new term query
func NewTermQuery(field string, value interface{}) *TermQuery {
	return &TermQuery{
		Field: field,
		Value: value,
	}
}

// NewRangeQuery creates a new range query
func NewRangeQuery(field string) *RangeQuery {
	return &RangeQuery{
		Field: field,
	}
}

// NewBoolQuery creates a new boolean query
func NewBoolQuery() *BoolQuery {
	return &BoolQuery{}
}

// AddMust adds a must query to a boolean query
func (q *BoolQuery) AddMust(query Query) *BoolQuery {
	q.Must = append(q.Must, query)
	return q
}

// AddMustNot adds a must not query to a boolean query
func (q *BoolQuery) AddMustNot(query Query) *BoolQuery {
	q.MustNot = append(q.MustNot, query)
	return q
}

// AddShould adds a should query to a boolean query
func (q *BoolQuery) AddShould(query Query) *BoolQuery {
	q.Should = append(q.Should, query)
	return q
}

// SetMinimumShouldMatch sets the minimum should match value
func (q *BoolQuery) SetMinimumShouldMatch(min int) *BoolQuery {
	q.MinimumShouldMatch = min
	return q
}

// NewExistsQuery creates a new exists query
func NewExistsQuery(field string) *ExistsQuery {
	return &ExistsQuery{
		Field: field,
	}
}

// NewPrefixQuery creates a new prefix query
func NewPrefixQuery(field, prefix string) *PrefixQuery {
	return &PrefixQuery{
		Field:  field,
		Prefix: prefix,
	}
}

// NewWildcardQuery creates a new wildcard query
func NewWildcardQuery(field, wildcard string) *WildcardQuery {
	return &WildcardQuery{
		Field:    field,
		Wildcard: wildcard,
	}
}

// NewGeoDistanceQuery creates a new geo distance query
func NewGeoDistanceQuery(field, distance string, lat, lon float64) *GeoDistanceQuery {
	return &GeoDistanceQuery{
		Field:    field,
		Distance: distance,
		Lat:      lat,
		Lon:      lon,
	}
}

// WithGT adds a greater than condition to a range query
func (q *RangeQuery) WithGT(value interface{}) *RangeQuery {
	q.GT = value
	return q
}

// WithGTE adds a greater than or equal condition to a range query
func (q *RangeQuery) WithGTE(value interface{}) *RangeQuery {
	q.GTE = value
	return q
}

// WithLT adds a less than condition to a range query
func (q *RangeQuery) WithLT(value interface{}) *RangeQuery {
	q.LT = value
	return q
}

// WithLTE adds a less than or equal condition to a range query
func (q *RangeQuery) WithLTE(value interface{}) *RangeQuery {
	q.LTE = value
	return q
}

// WithFormat adds a format to a range query
func (q *RangeQuery) WithFormat(format string) *RangeQuery {
	q.Format = format
	return q
}

// WithTimeZone adds a time zone to a range query
func (q *RangeQuery) WithTimeZone(timeZone string) *RangeQuery {
	q.TimeZone = timeZone
	return q
}

// WithRelation adds a relation to a range query
func (q *RangeQuery) WithRelation(relation string) *RangeQuery {
	q.Relation = relation
	return q
}

// QueryParser parses a query from a string
type QueryParser struct{}

// ParseQuery parses a simple query string into a Query object
func (p *QueryParser) ParseQuery(queryString string) (Query, error) {
	queryString = strings.TrimSpace(queryString)

	// Check for field:value syntax
	if strings.Contains(queryString, ":") {
		parts := strings.SplitN(queryString, ":", 2)
		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Check for range queries
		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			return p.parseRangeQuery(field, value)
		}

		// Check for wildcard queries
		if strings.Contains(value, "*") || strings.Contains(value, "?") {
			return NewWildcardQuery(field, value), nil
		}

		// Default to term query
		return NewTermQuery(field, value), nil
	}

	// Default to simple query
	return NewSimpleQuery(queryString), nil
}

// parseRangeQuery parses a range query from a string like "created:[2020-01-01 TO 2020-12-31]"
func (p *QueryParser) parseRangeQuery(field, value string) (Query, error) {
	// Remove brackets
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	// Split on TO
	if !strings.Contains(value, " TO ") {
		return nil, fmt.Errorf("invalid range query syntax: %s", value)
	}

	parts := strings.Split(value, " TO ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range query syntax: %s", value)
	}

	from := strings.TrimSpace(parts[0])
	to := strings.TrimSpace(parts[1])

	rangeQuery := NewRangeQuery(field)

	if from != "*" {
		rangeQuery.WithGTE(from)
	}

	if to != "*" {
		rangeQuery.WithLTE(to)
	}

	return rangeQuery, nil
}
