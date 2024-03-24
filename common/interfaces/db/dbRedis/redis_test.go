package dbRedis

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db"
	"github.com/Scalingo/sclng-backend-test-v1/common/util"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
)

const defaultPort = "6379"

var port string
var redisService *DBServiceRedis

// TestMain is the entry point for the test suite
// Setup Redis
// Run Tests
// Gracefully shutdown Redis
func TestMain(m *testing.M) {
	var err error

	var dockerPool *dockertest.Pool
	var resource *dockertest.Resource

	log := logger.Default().WithField("test", serviceName)

	// Fatal is a helper function to log an error and exit the program
	fatal := func(format string, v ...interface{}) {
		if err := dockerPool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %v", err)
		}
		log.Fatalf(format, v...)
	}

	// Setup Redis
	log.Info("Starting redis container...")

	dockerPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err = dockerPool.Run("redis/redis-stack-server", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	port = resource.GetPort(defaultPort + "/tcp")

	// Wait for the redis server to be available
	if waitErr := util.WaitForServiceOnPort(
		log, context.Background().Done(), serviceName, "localhost:"+port, 10*time.Second,
	); waitErr != nil {
		fatal("Could not connect to docker: %s", waitErr)
	}

	// Create a new DBServiceRedis instance & Start it
	redisService, err = NewDBServiceRedis(logrus.New(), ":"+port, "test")
	if err != nil {
		fatal("NewDBServiceRedis() error = %v", err)
		return
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	if err = redisService.Start(ctx, &wg); err != nil {
		fatal("Start() error = %v", err)
	}

	// Run the tests
	returnCode := m.Run()

	// Once finished cancel the servicesContext to shut down the redis service
	cancel()

	// Block while the redis service is running
	wg.Wait()

	// Cleanup
	if err := dockerPool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(returnCode)

}

func TestDBServiceRedis_SetRepoItem_SetLanguages_GetItem(t *testing.T) {
	testKey := t.Name()
	if err := redisService.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	db.SetRepoList_SetLanguages_GetItem(t, redisService, testKey)
}

func TestSetRepoList_PreserveLanguages(t *testing.T) {
	testKey := t.Name()
	if err := redisService.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	db.SetRepoList_PreserveLanguages(t, redisService, testKey)
}
