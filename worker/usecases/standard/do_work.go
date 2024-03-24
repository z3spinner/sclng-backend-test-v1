package standard

import (
	"context"
	"runtime"

	"github.com/Scalingo/sclng-backend-test-v1/common/entities"
	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Standard) doWork(ctx context.Context) error {
	s.log.Info("Working")

	// **********************************************************************
	// 1. Fetch the latest 100 repositories
	//    - If the fetch is rate limited, wait until the rate limit is reset
	// **********************************************************************

	var repoList entities.RepoList
	err := s.retryOrWait(
		ctx, func(ctx context.Context) error {

			// This is where the initial request to get the latest 100 repositories is made

			var err error
			repoList, err = s.fetch.GetRepoList(ctx)
			if err != nil {
				s.logFetchError(s.log, err, "repoList")
				return err
			}
			s.log.Info("fetched repoList")

			err = s.db.SetRepoList(ctx, repoList)
			if err != nil {
				return errors.Wrap(err, "error storing repoList in db")
			}
			s.log.Info("stored repoList")
			return nil
		},
	)
	if err != nil {
		s.log.Errorf("error fetching repoList: %v", err)
	}

	// **********************************************************************
	// 2. Fetch the languages for each repository
	//    - If the fetch is rate limited, wait until the rate limit is reset
	//    - Fetch the languages for each repository in parallel
	// **********************************************************************

	// Set the maximum number of parallel fetches
	maxParallel := runtime.NumCPU()

	// Create a channel to limit the number of parallel fetches
	//  - runningWorkers: the number of workers currently running
	//  - availableWorkers: the number of workers available to run (exists for logging purposes)
	// When a worker is running, its ID is moved availableWorkers -> -> runningWorkers
	// When a worker completes, its ID is moved runningWorkers <- <- availableWorkers
	runningWorkers := make(chan int, maxParallel)
	availableWorkers := make(chan int, maxParallel)
	for i := 0; i < maxParallel; i++ {
		availableWorkers <- i
	}

	// Create a channel to receive errors when each worker completes
	errs := make(chan error, len(repoList))

	// Fetch the languages for each repository
	for _, repo := range repoList {

		// Move a worker from the availableWorkers channel to the runningWorkers channel
		workerID := <-availableWorkers
		runningWorkers <- workerID

		go func(repo entities.RepoItem, currentWorkerID int) {
			defer func() { availableWorkers <- <-runningWorkers }()

			err := s.retryOrWait(
				context.WithValue(ctx, "workerID", currentWorkerID),
				func(ctx context.Context) error {

					log := s.log.WithFields(
						map[string]interface{}{
							"repo":     repo.Name,
							"workerID": currentWorkerID,
						},
					)
					// This is where the request to get the languages for a repository is made
					// This happens in parallel
					newLangs, err := s.fetch.GetRepoLanguages(ctx, repo.LanguagesURL)
					if err != nil {
						s.logFetchError(log, err, "languages")
						return err
					}
					log.Info("fetched languages")

					err = s.db.SetRepoItemLanguages(ctx, repo.ID, newLangs)
					if err != nil {
						return errors.Wrap(err, "error storing languages in db")
					}
					log.Info("stored languages")

					return nil
				},
			)
			if err != nil {
				s.log.Errorf("error fetching languages: %v", err)
			}
			errs <- err
		}(repo, workerID)
	}

	// Wait for all fetches to complete
	for i := 0; i < len(repoList); i++ {
		select {
		case _ = <-errs:
			//if err != nil {
			//	s.log.Errorf("error fetching languages: %v", err)
			//}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	s.log.Info("Work complete")

	return nil
}

// logFetchError logs the error from a fetch request
// It logs the error message and any additional information
// It takes the parameters:
//   - log: a logger to log the error. (This is included to allow for additional context to be added to the log)
//   - err: the error to log
//   - name: the name of fetch call
func (s *Standard) logFetchError(log logrus.FieldLogger, err error, name string) {
	if errors.Is(err, fetcher.ErrRateLimited) {
		log.WithField("endpoint", name).
			Warnf(
				"rate limit exceeded",
			)
	} else if errors.Is(err, fetcher.ErrRequestTimeout) {
		log.WithField("endpoint", name).Warn("request timed out")
	}
}
