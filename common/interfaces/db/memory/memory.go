package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const serviceName = "DBServiceMemory"

// Custom map key types
type repoKey string

// expiringValue is struct to hold data with an optional expiresAt time
type expiringValue struct {
	value     []byte
	expiresAt time.Time
}

// DBServiceMemory is a memory db service
type DBServiceMemory struct {
	log   logrus.FieldLogger
	mutex *sync.Mutex

	dataItems map[repoKey]entities.RepoItem
}

// getRepoKey returns the key for a repo entry
func getRepoKey(repoID int64) repoKey {
	return repoKey(fmt.Sprintf("repo:%d", repoID))
}

// GetAvgNumForksPerRepoByLanguage returns the average number of forks per repo by language
func (c *DBServiceMemory) GetAvgNumForksPerRepoByLanguage(ctx context.Context) (map[string]float32, error) {
	//TODO implement me
	//  NOTE: I did not have time to implement this
	panic("implement me")
}

// GetNumReposByLanguage returns the number of repos by language
func (c *DBServiceMemory) GetNumReposByLanguage(ctx context.Context) (map[string]int, error) {
	//TODO implement me
	//  NOTE: I did not have time to implement this
	panic("implement me")
}

// GetAvgNumOpenIssuesByLanguage returns the average number of open issues by language
func (c *DBServiceMemory) GetAvgNumOpenIssuesByLanguage(ctx context.Context) (map[string]float32, error) {
	//TODO implement me
	//  NOTE: I did not have time to implement this
	panic("implement me")
}

// GetAvgSizeByLanguage returns the average size by language
func (c *DBServiceMemory) GetAvgSizeByLanguage(ctx context.Context) (map[string]float32, error) {
	//TODO implement me
	//  NOTE: I did not have time to implement this
	panic("implement me")
}

// SetRepoItemLanguages sets the languages for a repo item
func (c *DBServiceMemory) SetRepoItemLanguages(ctx context.Context, repoID int64, langs entities.Languages) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	keyRepo := getRepoKey(repoID)
	item, ok := c.dataItems[keyRepo]
	if !ok {
		return db.ErrNotFound
	}
	item.Languages = langs
	c.dataItems[keyRepo] = item

	return nil
}

// GetRepoItem returns a repo item
func (c *DBServiceMemory) GetRepoItem(ctx context.Context, repoID int64) (entities.RepoItem, error) {

	keyItem := getRepoKey(repoID)

	// Get the item
	if item, ok := c.dataItems[keyItem]; ok {
		return item, nil
	}
	return entities.RepoItem{}, db.ErrNotFound
}

// setRepoItem sets a repo item
func (c *DBServiceMemory) setRepoItem(ctx context.Context, item entities.RepoItem) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	keyItem := getRepoKey(item.ID)
	c.dataItems[keyItem] = item

	return nil
}

// SetRepoList sets a list of repo items
func (c *DBServiceMemory) SetRepoList(ctx context.Context, list entities.RepoList) error {

	// Keep a copy of the existing item IDs
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
		delete(c.dataItems, getRepoKey(id))
	}

	return nil
}

// GetRepoList returns a list of repo items
func (c *DBServiceMemory) GetRepoList(ctx context.Context, filters db.GetRepoListFilters) (entities.RepoList, error) {
	list := entities.RepoList{}
	for _, item := range c.dataItems {
		list = append(list, item)
	}
	return list, nil
}

var _ db.Service = (*DBServiceMemory)(nil)
