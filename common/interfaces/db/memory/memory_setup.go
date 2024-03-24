package memory

import (
	"sync"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
	"github.com/sirupsen/logrus"
)

// New creates a new memory db service
func New(log logrus.FieldLogger) (*DBServiceMemory, error) {

	log = log.WithFields(
		logrus.Fields{
			"service": serviceName,
		},
	)

	return &DBServiceMemory{
		log:       log,
		mutex:     &sync.Mutex{},
		dataItems: map[repoKey]entities.RepoItem{},
	}, nil
}

// Reset resets the db
func (c *DBServiceMemory) Reset() {
	c.dataItems = map[repoKey]entities.RepoItem{}
}
