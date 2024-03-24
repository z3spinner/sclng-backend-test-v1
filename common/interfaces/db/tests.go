package db

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
)

var SetRepoList_SetLanguages_GetItem = setRepoList_SetLanguages_GetItem
var SetRepoList_PreserveLanguages = setRepoList_PreserveLanguages

func setRepoList_SetLanguages_GetItem(t *testing.T, dbService Service, testKey string) {

	item := entities.RepoItem{
		ID:              1,
		Name:            "2",
		FullName:        "3",
		Owner:           "4",
		HTMLUrl:         "5",
		Description:     "6",
		LanguagesURL:    "7",
		CreatedAt:       time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
		Size:            8,
		Language:        "9",
		Languages:       nil,
		LicenseName:     "10",
		ForksCount:      11,
		OpenIssuesCount: 12,
		WatchersCount:   13,
		AllowForking:    false,
		HasIssues:       false,
		HasProjects:     false,
		HasDownloads:    false,
		HasWiki:         false,
		HasPages:        false,
		HasDiscussions:  false,
	}

	err := dbService.SetRepoList(context.Background(), entities.RepoList{item})
	if err != nil {
		t.Errorf("setRepoItem() error = %v", err)
		return
	}

	// GET
	val, err := dbService.GetRepoItem(context.Background(), item.ID)
	if err != nil {
		t.Errorf("GetRepoItem() error = %v", err)
		return
	}
	if !reflect.DeepEqual(val, item) {
		t.Errorf("GetRepoItem()\ngot =  %v\nwant = %v", val, item)
		return
	}

	// SET The languages
	languages := entities.Languages{
		"A": 1,
		"B": 2,
	}
	err = dbService.SetRepoItemLanguages(context.Background(), item.ID, languages)
	if err != nil {
		t.Errorf("SetRepoItemLanguages() error = %v", err)
		return
	}

	item.Languages = languages

	// GET
	val, err = dbService.GetRepoItem(context.Background(), item.ID)
	if err != nil {
		t.Errorf("GetRepoItem() error = %v", err)
		return
	}
	if !reflect.DeepEqual(val, item) {
		t.Errorf("GetRepoItem() val = %v, want %v", val, item)
		return
	}

}

// setRepoList_PreserveLanguages checks that on a repeated cycle (GetRepoList, GetLanguages, GetRepoList) the languages are not lost in the process.
func setRepoList_PreserveLanguages(t *testing.T, dbService Service, testKey string) {
	itemV1 := entities.RepoItem{
		ID:           1,
		Name:         "2",
		FullName:     "3",
		Owner:        "4",
		HTMLUrl:      "5",
		Description:  "6",
		LanguagesURL: "7",
		CreatedAt:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
		Size:         8,
		Language:     "9",
		Languages: map[string]int64{
			"A": 1,
			"B": 2,
		},
		LicenseName:     "10",
		ForksCount:      11,
		OpenIssuesCount: 12,
		WatchersCount:   13,
		AllowForking:    false,
		HasIssues:       false,
		HasProjects:     false,
		HasDownloads:    false,
		HasWiki:         false,
		HasPages:        false,
		HasDiscussions:  false,
	}

	err := dbService.SetRepoList(context.Background(), entities.RepoList{itemV1})
	if err != nil {
		t.Errorf("setRepoItem() error = %v", err)
		return
	}

	// GET
	val, err := dbService.GetRepoItem(context.Background(), itemV1.ID)
	if err != nil {
		t.Errorf("GetRepoItem() error = %v", err)
		return
	}
	if !reflect.DeepEqual(val, itemV1) {
		t.Errorf("GetRepoItem() val = %v, want %v", val, itemV1)
		return
	}

	// Overwrite the repo item with a new one that has no languages
	itemV2 := itemV1
	itemV2.Languages = nil

	err = dbService.SetRepoList(context.Background(), entities.RepoList{itemV2})
	if err != nil {
		t.Errorf("setRepoItem() error = %v", err)
		return
	}

	// GET the item again (it should still have the languages)
	val, err = dbService.GetRepoItem(context.Background(), itemV1.ID)
	if err != nil {
		t.Errorf("GetRepoItem() error = %v", err)
		return
	}
	if !reflect.DeepEqual(val, itemV1) {
		t.Errorf("GetRepoItem() val = %v, want %v", val, itemV1)
		return
	}

}
