package memory

import (
	"os"
	"testing"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db"
	"github.com/sirupsen/logrus"
)

var memoryService *DBServiceMemory

// TestMain is the entry point for the test suite
// Setup DB
// Run Tests
// Cleanup
func TestMain(m *testing.M) {
	var err error

	log := logger.Default().WithField("test", serviceName)

	// Create a new DBServiceMemory instance & Start it
	memoryService, err = New(logrus.New())
	if err != nil {
		log.Fatalf("New() error = %v", err)
		return
	}

	// Run the tests
	returnCode := m.Run()

	memoryService.Reset()

	os.Exit(returnCode)

}

func TestDBServiceMemory_SetRepoItem_SetLanguages_GetItem(t *testing.T) {
	testKey := t.Name()
	memoryService.Reset()
	db.SetRepoList_SetLanguages_GetItem(t, memoryService, testKey)
}

func TestSetRepoList_PreserveLanguages(t *testing.T) {
	testKey := t.Name()
	memoryService.Reset()
	db.SetRepoList_PreserveLanguages(t, memoryService, testKey)
}
