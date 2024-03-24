package usecases

import (
	"context"
	"sync"
)

type Usecases interface {
	RunWorker(ctx context.Context, wg *sync.WaitGroup) error
}
