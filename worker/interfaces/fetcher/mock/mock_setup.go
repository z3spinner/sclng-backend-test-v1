package fetcherMock

import (
	"os"
	"sync"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/worker/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// New creates a new mock Fetcher service
func New(log logrus.FieldLogger, dataDir string, cfg *config.Config) (*FetcherMock, error) {

	log = log.WithFields(
		logrus.Fields{
			"service": serviceName,
		},
	)

	// ensure dataDir has a trailing slash
	if dataDir[len(dataDir)-1] != '/' {
		dataDir += "/"
	}

	// check if dataDir exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.WithField("dataDir", dataDir).Warn("dataDir does not exist")
		wd, _ := os.Getwd()
		return nil, errors.Wrapf(err, "dataDir does not exist: %s (workingDir: %s)", dataDir, wd)
	}

	return &FetcherMock{
		log:             log,
		dataDir:         dataDir,
		cfg:             cfg,
		mockRateLimiter: newRateLimiter(cfg.MockRateLimit, time.Duration(cfg.MockRateLimitWindowSeconds)*time.Second),
		fileCache:       make(map[string][]byte),
		fileCacheMutex:  sync.Mutex{},
	}, nil
}
