package fetcher

import (
	"context"
	"testing"
)

var GetLatest100Test = getLatest100Test
var GetRepoLanguagesTest = getRepoLanguagesTest

func getLatest100Test(t *testing.T, fetcher Service) {
	ctx := context.Background()

	repos, err := fetcher.GetRepoList(ctx)
	if err != nil {
		t.Errorf("GetRepoList() error = %v", err)
		return
	}

	if len(repos) != 100 {
		t.Errorf("GetRepoList() got %d repos, want 100", len(repos))
		return
	}
}

func getRepoLanguagesTest(t *testing.T, fetcher Service) {
	ctx := context.Background()

	repos, err := fetcher.GetRepoList(ctx)
	if err != nil {
		t.Errorf("GetRepoList() error = %v", err)
		return
	}

	if len(repos) != 100 {
		t.Errorf("GetRepoList() got %d repos, want 100", len(repos))
		return
	}

	langs, err := fetcher.GetRepoLanguages(ctx, repos[0].LanguagesURL)
	if err != nil {
		t.Errorf("GetRepoLanguages() error = %v", err)
		return
	}

	if langs == nil {
		t.Errorf("GetRepoLanguages() got nil, want map")
		return
	}
	if len(langs) < 2 {
		t.Errorf("GetRepoLanguages()languages, want >=2,  got %d", len(langs))
		return
	}

	// Note extra tests exclusively in the "mock" implementation
}
