package fetcherMock

import (
	"context"
	"os"
	"testing"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/worker/config"
	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
)

var mockService *FetcherMock

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
	cfg.FetchTimeoutSeconds = 10
	cfg.MockFetcherAvgRequestSeconds = 0

	// Create a new FetcherMock instance & Start it
	mockService, err = New(log, "data", cfg)
	if err != nil {
		log.Fatalf("New() error = %v", err)
		return
	}

	// Run the tests
	returnCode := m.Run()

	// Cleanup
	mockService = nil // Leave it to the garbage collector

	os.Exit(returnCode)

}

func TestFetcherMock_GetLatest100(t *testing.T) {
	fetcher.GetLatest100Test(t, mockService)
}

func TestFetcherMock_GetRepoLanguages(t *testing.T) {
	fetcher.GetRepoLanguagesTest(t, mockService)
}

func TestFetcherMock_GetRepoLanguagesCheckLines(t *testing.T) {
	ctx := context.Background()

	repos, err := mockService.GetRepoList(ctx)
	if err != nil {
		t.Errorf("GetRepoList() error = %v", err)
		return
	}

	if len(repos) != 100 {
		t.Errorf("GetRepoList() got %d repos, want 100", len(repos))
		return
	}

	//{
	//  "Python": 1215553,
	//  "Smarty": 130
	//}
	langs, err := mockService.GetRepoLanguages(ctx, "languages")
	if err != nil {
		t.Errorf("GetRepoLanguages() error = %v", err)
		return
	}
	if langs == nil {
		t.Errorf("GetRepoLanguages() got nil, want map")
		return
	}
	if len(langs) != 2 {
		t.Errorf("GetRepoLanguages() got %d languages, want 2", len(langs))
		return
	}
	if langs["Python"] != 1215553 {
		t.Errorf("GetRepoLanguages() langs[\"Python\"] = %d, want 1215553", langs["Python"])
		return
	}
	if langs["Smarty"] != 130 {
		t.Errorf("GetRepoLanguages() langs[\"Smarty\"] = %d, want 130", langs["Smarty"])
		return
	}
}
