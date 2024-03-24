package dbRedis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// DBServiceRedis is a redis db service
type DBServiceRedis struct {
	pool      *redis.Client
	log       logrus.FieldLogger
	keyPrefix string
	hostPort  string
}

// CreateIndexes creates the indexes for the redis db
func (c *DBServiceRedis) CreateIndexes(ctx context.Context) error {
	// Create the indexes
	//"FT.CREATE idx:repo ON JSON PREFIX 1 repo: SCHEMA $.id as id NUMERIC $.name as name TEXT $.language as language TEXT $.all_languages as all_languages TAG $.license as license TEXT $.size as size NUMERIC $.watchers_count as watchers_count NUMERIC $.forks_count as forks_count NUMERIC $.allow_forking as allow_forking TAG $.open_issues_count as open_issues_count NUMERIC"

	err := c.pool.Do(
		ctx, "FT.CREATE", "idx:repo", "ON", "JSON", "PREFIX", "1", "repo:", "SCHEMA",
		"$.id", "as", "id", "NUMERIC",
		"$.name", "as", "name", "TEXT",
		"$.language", "as", "language", "TEXT",
		"$.all_languages", "as", "all_languages", "TAG",
		"$.license", "as", "license", "TEXT",
		"$.size", "as", "size", "NUMERIC",
		"$.watchers_count", "as", "watchers_count", "NUMERIC",
		"$.forks_count", "as", "forks_count", "NUMERIC",
		"$.allow_forking", "as", "allow_forking", "TAG",
		"$.open_issues_count", "as", "open_issues_count", "NUMERIC",
	).Err()
	if err != nil && err.Error() != "Index already exists" {
		return errors.Wrap(err, "Could not create index")
	}
	return nil
}

// setRepoItem sets a repo item in the db
func (c *DBServiceRedis) setRepoItem(ctx context.Context, item entities.RepoItem) error {

	// Convert to the redis interface type
	doc, err := ConvertRepoItemE2I(item)
	if err != nil {
		return errors.Wrap(err, "Error converting repo")
	}

	key := doc.getKey()

	// Marshal the JSON document into a string
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return errors.Wrap(err, "Error marshaling JSON:")
	}

	// Store the JSON document in Redis
	err = c.pool.JSONSet(ctx, string(key), "$", jsonData).Err()
	if err != nil {
		return errors.Wrap(err, "Error storing document in Redis:")
	}

	return nil
}

// SetRepoItemLanguages sets the languages field in the repo document
func (c *DBServiceRedis) SetRepoItemLanguages(ctx context.Context, repoID int64, langs entities.Languages) error {

	key := getRepoKey(repoID)

	iLangs, err := ConvertLanguagesE2I(langs)
	if err != nil {
		return errors.Wrap(err, "Error converting languages")
	}

	// Marshal the JSON document into a string
	jsonData, err := json.Marshal(iLangs)
	if err != nil {
		return errors.Wrap(err, "Error marshaling JSON:")
	}

	// Store the JSON document in Redis
	err = c.pool.JSONMerge(ctx, string(key), "$.languages", string(jsonData)).Err()
	if err != nil {
		return errors.Wrap(err, "Error storing document in Redis:")
	}

	// Set the all_languages field (combines Language and languages fields)
	err = c.setAllLanguages(ctx, repoID, langs)
	if err != nil {
		return errors.Wrap(err, "Error setting all languages")
	}

	return nil
}

// SetAllLanguages sets the all_languages field in the repo document
// Combine repo.language field and the keys of repo.languages field into a single array of all_languages for searching
func (c *DBServiceRedis) setAllLanguages(ctx context.Context, repoID int64, langs entities.Languages) error {

	key := getRepoKey(repoID)

	// Get the repo document
	repo, err := c.GetRepoItem(ctx, repoID)
	if err != nil {
		return errors.Wrap(err, "Error getting repo")
	}

	// Combine the languages into a single array
	allLanguages := append(repo.Languages.Strings(), repo.Language)

	// Marshal the JSON document into a string
	jsonData, err := json.Marshal(allLanguages)
	if err != nil {
		return errors.Wrap(err, "Error marshaling JSON:")
	}

	// Store the JSON document in Redis
	err = c.pool.JSONMerge(ctx, string(key), "$.all_languages", string(jsonData)).Err()
	if err != nil {
		return errors.Wrap(err, "Error storing document in Redis:")
	}

	return nil
}

// SetRepoList sets a list of repo items in the db
func (c *DBServiceRedis) SetRepoList(ctx context.Context, list entities.RepoList) error {

	// Keep a copy of the existing items
	existingItems, err := c.GetRepoList(ctx, db.GetRepoListFilters{})
	if err != nil {
		return errors.Wrap(err, "could not get existing items")
	}
	existingItemIDs := make(map[int64]entities.RepoItem)
	for _, item := range existingItems {
		existingItemIDs[item.ID] = item
	}

	// Iterate through the new list
	// As we iterate, we will remove the IDs from the existingItems map.
	// This will leave us with a list of IDs that we can delete
	for _, item := range list {

		// Copy the languages from the existing item to the new one (so we don't lose the data)
		if existingItemIDs[item.ID].Languages != nil {
			item.Languages = existingItemIDs[item.ID].Languages
		}

		// Delete the item from the existingItems map
		delete(existingItemIDs, item.ID)

		// Set the item in the DB
		err := c.setRepoItem(ctx, item)
		if err != nil {
			return errors.Wrap(err, "error storing repoList item")
		}
	}

	// existingItems now contains the IDs that are no longer in the list. We can delete them
	for id := range existingItemIDs {
		key := getRepoKey(id)
		err = c.pool.JSONDel(ctx, string(key), "$").Err()
		if err != nil {
			return errors.Wrap(err, "Error deleting document in Redis:")
		}
	}

	return nil
}

// GetRepoList retrieves a list of repo items from the db
func (c *DBServiceRedis) GetRepoList(ctx context.Context, filters db.GetRepoListFilters) (entities.RepoList, error) {
	//Get all the repos from redis

	// Build the query
	query := buildQueryFromFilters(filters)

	// Make the search for the keys
	// TODO: Factor the call to c.pool.Do out to a function for making a search for keys
	var err error
	var list struct {
		Results []struct {
			ID string
		}
	}
	res, err := c.pool.Do(ctx, "FT.SEARCH", "idx:repo", query, "LIMIT", 0, 100, "NOCONTENT").Result()
	if err != nil {
		return nil, errors.Wrap(err, "Error searching for repos")
	}

	// Decode the returned data
	err = mapstructure.Decode(res, &list)
	if err != nil {
		return nil, err
	}

	// No results
	if len(list.Results) == 0 {
		return entities.RepoList{}, nil
	}

	// Fetch the full json using the keys
	keys := make([]string, len(list.Results))
	for i := 0; i < len(keys); i++ {
		keys[i] = list.Results[i].ID
	}

	jsonList, err := c.pool.JSONMGet(ctx, "$", keys...).Result()
	if err != nil {
		return nil, err
	}

	repoList := make(RepoList, len(jsonList))
	for i := 0; i < len(jsonList); i++ {
		var doc []RepoItem
		data := jsonList[i].(string)
		err = json.Unmarshal([]byte(data), &doc)
		if err != nil {
			return nil, errors.Wrap(err, "Error unmarshaling document")
		}
		repoList[i] = doc[0]
	}

	// Return the RepoList
	return ConvertRepoListI2E(repoList), nil
}

// buildQueryFromFilters builds a query string from the filters
func buildQueryFromFilters(filters db.GetRepoListFilters) string {

	// Build the filters
	var filterParams []string
	if filters.Name != nil && *filters.Name != "" {
		filterParams = append(filterParams, fmt.Sprintf("@name: *%s*", *filters.Name))
	}

	if filters.Language != nil && *filters.Language != "" {
		filterParams = append(filterParams, fmt.Sprintf("@all_languages: {*%s*}", *filters.Language))
	}

	if filters.License != nil && *filters.License != "" {
		filterParams = append(filterParams, fmt.Sprintf("@license: *%s*", *filters.License))
	}

	if filters.AllowForking != nil {
		filterParams = append(filterParams, fmt.Sprintf("@allow_forking:{%v}", *filters.AllowForking))
	}

	if filters.HasOpenIssues != nil {
		if *filters.HasOpenIssues {
			filterParams = append(filterParams, "@open_issues_count:[1 +inf]")
		} else {
			filterParams = append(filterParams, "@open_issues_count:[0 0]")
		}
	}

	filter := "*"
	if len(filterParams) > 0 {
		filter = strings.Join(filterParams, " ")
	}

	return filter
}

// GetRepoItem retrieves a repo item from the db
func (c *DBServiceRedis) GetRepoItem(ctx context.Context, repoID int64) (entities.RepoItem, error) {
	key := getRepoKey(repoID)

	jsonData, err := c.pool.JSONGet(ctx, string(key)).Result()
	if err != nil {
		return entities.RepoItem{}, err
	}

	var doc RepoItem
	err = json.Unmarshal([]byte(jsonData), &doc)
	if err != nil {
		return entities.RepoItem{}, errors.Wrap(err, "Error unmarshaling document")
	}

	return ConvertRepoItemI2E(doc)
}

// GetAvgNumForksPerRepoByLanguage returns the average number of forks per repo by language
func (c *DBServiceRedis) GetAvgNumForksPerRepoByLanguage(ctx context.Context) (map[string]float32, error) {

	res, err := c.pool.Do(
		ctx, "FT.AGGREGATE", "idx:repo", "*", "LOAD", "1", "@forks_count", "GROUPBY", "1", "@language",
		"REDUCE", "AVG", "1", "@forks_count", "AS", "count",
		"LIMIT", "0", "1000",
	).Result()
	if err != nil {
		return nil, errors.Wrap(err, "Error searching for repos")
	}

	return decodeAggregateResultFloat32(res)
}

// GetNumReposByLanguage returns the number of repos by language
func (c *DBServiceRedis) GetNumReposByLanguage(ctx context.Context) (map[string]int, error) {
	res, err := c.pool.Do(
		ctx, "FT.AGGREGATE", "idx:repo", "*", "LOAD", "1", "@name", "GROUPBY", "1", "@language",
		"REDUCE", "COUNT_DISTINCT", "1", "@name", "AS", "count",
		"LIMIT", "0", "1000",
	).Result()
	if err != nil {
		return nil, errors.Wrap(err, "Error searching for repos")
	}

	return decodeAggregateResultInt(res)
}

// GetAvgNumOpenIssuesByLanguage returns the average number of open issues by language
func (c *DBServiceRedis) GetAvgNumOpenIssuesByLanguage(ctx context.Context) (map[string]float32, error) {
	res, err := c.pool.Do(
		ctx, "FT.AGGREGATE", "idx:repo", "*", "LOAD", "1", "@open_issues_count", "GROUPBY", "1", "@language",
		"REDUCE", "AVG", "1", "@open_issues_count", "AS", "count",
		"LIMIT", "0", "1000",
	).Result()
	if err != nil {
		return nil, errors.Wrap(err, "Error searching for repos")
	}

	return decodeAggregateResultFloat32(res)
}

// GetAvgSizeByLanguage returns the average size by language
func (c *DBServiceRedis) GetAvgSizeByLanguage(ctx context.Context) (map[string]float32, error) {
	res, err := c.pool.Do(
		ctx, "FT.AGGREGATE", "idx:repo", "*", "LOAD", "1", "@size", "GROUPBY", "1", "@language",
		"REDUCE", "AVG", "1", "@size", "AS", "count",
		"LIMIT", "0", "1000",
	).Result()
	if err != nil {
		return nil, errors.Wrap(err, "Error searching for repos")
	}

	return decodeAggregateResultFloat32(res)
}

// decodeAggregateResultString decodes the result of an aggregate query into a map of strings
func decodeAggregateResultString(in interface{}) (map[string]string, error) {
	var list struct {
		Results []struct {
			Extra_Attributes struct {
				Language string
				Count    string
			}
		}
	}

	// Decode the returned data
	err := mapstructure.Decode(in, &list)
	if err != nil {
		return nil, err
	}

	out := map[string]string{}
	for _, row := range list.Results {
		out[row.Extra_Attributes.Language] = row.Extra_Attributes.Count
	}

	return out, nil
}

// decodeAggregateResultFloat32 decodes the result of an aggregate query into a map of float32
func decodeAggregateResultFloat32(in interface{}) (map[string]float32, error) {

	strs, err := decodeAggregateResultString(in)
	if err != nil {
		return nil, err
	}

	out := map[string]float32{}
	for k, v := range strs {
		avg, _ := strconv.ParseFloat(v, 32)
		out[k] = float32(avg)
	}

	return out, nil
}

// decodeAggregateResultInt decodes the result of an aggregate query into a map of int
func decodeAggregateResultInt(in interface{}) (map[string]int, error) {

	strs, err := decodeAggregateResultString(in)
	if err != nil {
		return nil, err
	}

	out := map[string]int{}
	for k, v := range strs {
		avg, _ := strconv.ParseInt(v, 10, 32)
		out[k] = int(avg)
	}

	return out, nil
}

var _ db.Service = (*DBServiceRedis)(nil)
