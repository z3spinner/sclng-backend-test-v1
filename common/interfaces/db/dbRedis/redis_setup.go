package dbRedis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/common/util"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const serviceName = "DBServiceRedis"

// NewDBServiceRedis creates a new redis db service
func NewDBServiceRedis(log logrus.FieldLogger, hostPort, keyPrefix string) (*DBServiceRedis, error) {

	log = log.WithFields(
		logrus.Fields{
			"service":  serviceName,
			"hostPort": hostPort,
		},
	)

	return &DBServiceRedis{
		log:       log,
		keyPrefix: keyPrefix,
		hostPort:  hostPort,
	}, nil
}

// Start starts the redis service in a goroutine
// Returns an error if the service fails to start
// Graceful shutdown when an interrupt signal is received from the OS
func (c *DBServiceRedis) Start(ctx context.Context, wg *sync.WaitGroup) error {

	if ctx == nil {
		return fmt.Errorf("parent context is required")
	}

	if wg == nil {
		return fmt.Errorf("wait group is required")
	}

	// Increment the wait group counter
	wg.Add(1)

	// Wait for the redis service to be available
	err := util.WaitForServiceOnPort(c.log, ctx.Done(), serviceName, c.hostPort, 10*time.Second)
	if err != nil {
		wg.Done()
		return errors.Wrap(err, "Could not dial redis")
	}

	// Create a new client connection
	c.pool = redis.NewClient(
		&redis.Options{
			Addr:     c.hostPort, // use default Addr
			Password: "",         // no password set
			DB:       0,          // use default DB
		},
	)

	// Ping the redis server to check the connection
	pong, err := c.pool.Ping(ctx).Result()
	if err != nil || pong != "PONG" {
		wg.Done()
		return errors.Wrap(err, "Could not ping redis")
	}

	// Create the indexes
	err = c.CreateIndexes(ctx)
	if err != nil {
		wg.Done()
		return errors.Wrap(err, "Could not create indexes")
	}

	c.log.Info("connected to redis")

	// Wait for the context to be cancelled in a goroutine, gracefully shutdown the service
	var routineErr error
	go func() {
		defer wg.Done()

		// Listen for the context to be cancelled. This blocks until the signal is received.
		<-ctx.Done()

		// Ask the redis pool to close all connections and cleanup.
		c.log.Info("graceful shutdown started")
		err := c.pool.Close()
		if err != nil {
			routineErr = errors.Wrap(err, "could not close connection pool")
		}
		c.log.Info("graceful shutdown complete")

	}()

	return routineErr
}

// Reset resets the db
func (c *DBServiceRedis) Reset() error {
	c.pool.FlushAll(context.Background())
	return c.CreateIndexes(context.Background())
}
