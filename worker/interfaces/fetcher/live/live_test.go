package fetcherLive

import (
	"os"
	"testing"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/worker/config"
	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
)

var liveService *FetcherLive

// TestMain is the entry point for the test suite
// Setup
// Run Tests
// Cleanup
func TestMain(m *testing.M) {
	var err error

	log := logger.Default().WithField("test", serviceName)
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config.New() error = %v", err)
		return
	}

	// Create a new FetcherLive instance & Start it
	liveService, err = New(log, cfg)
	if err != nil {
		log.Fatalf("New() error = %v", err)
		return
	}

	// Run the tests
	returnCode := m.Run()

	// Cleanup
	liveService = nil // Leave it to the garbage collector

	os.Exit(returnCode)

}

func TestFetcherLive_GetLatest100(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	fetcher.GetLatest100Test(t, liveService)
}

func TestFetcherLive_GetRepoLanguages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	fetcher.GetRepoLanguagesTest(t, liveService)
}
