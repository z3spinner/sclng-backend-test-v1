package webservice

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/apiServer/config"
	"github.com/Scalingo/sclng-backend-test-v1/apiServer/usecases/standard"
	"github.com/Scalingo/sclng-backend-test-v1/common/interfaces/db/memory"
	"github.com/sirupsen/logrus"
)

// TestNewWebservice tests the New function
// New should return an appropriate error, if the arguments are incorrect
func TestNewWebservice(t *testing.T) {

	ctx := context.Background()

	// Logger
	log := logger.Default()

	// Config
	cfg, err := config.New()
	if err != nil {
		t.Fatalf(`failed to create config: %v`, err)
	}

	// DB Service (memory
	db, err := memory.New(log)

	// Usecases Layer
	uc := standard.New(ctx, log, cfg, db)

	type args struct {
		log    logrus.FieldLogger
		config *config.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Correctly configured",
			args: args{
				log: log,
				config: &config.Config{
					APIServerPort: 5001,
				},
			},
			wantErr: false,
		},
		{
			name:    "Nil Everything",
			args:    args{},
			wantErr: true,
		},
		{
			name: "Nil Logger",
			args: args{
				config: &config.Config{
					APIServerPort: 5001,
				},
			},
			wantErr: true,
		},
		{
			name: "Nil Config",
			args: args{
				log: log,
			},
			wantErr: true,
		},
		{
			name: "Empty Config",
			args: args{
				log:    log,
				config: &config.Config{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := New(tt.args.log, tt.args.config, uc)
				if (err != nil) != tt.wantErr {
					t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// If we don't want an error, then we should check the result
				if !tt.wantErr {
					if got.log == nil || got.serverPort != tt.args.config.APIServerPort {
						t.Errorf("New() got = %v", got)
					}
				}
			},
		)
	}
}

// TestWebservice_Start tests the Start method of the Webservice
// Start method should return an appropriate error, if the arguments are incorrect
func TestWebservice_Start(t *testing.T) {

	ctx := context.Background()

	// Logger
	log := logger.Default()

	// Config
	cfg, err := config.New()
	if err != nil {
		t.Fatalf(`failed to create config: %v`, err)
	}

	// DB Service (memory
	db, err := memory.New(log)

	// Usecases Layer
	uc := standard.New(ctx, log, cfg, db)

	ws, err := New(
		log, &config.Config{
			APIServerPort: 5001,
		}, uc,
	)
	if err != nil {
		t.Fatalf(`failed to create webservice: %v`, err)
	}

	type fields struct {
		ws *Webservice
	}
	type args struct {
		parentCtx context.Context
		wg        *sync.WaitGroup
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Correctly configured",
			fields: fields{
				ws: ws,
			},
			args: args{
				parentCtx: context.Background(),
				wg:        &sync.WaitGroup{},
			},
			wantErr: false,
		},
		{
			name: "Nil Everything",
			fields: fields{
				ws: ws,
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "Nil WaitGroup",
			fields: fields{
				ws: ws,
			},
			args: args{
				parentCtx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "Nil ParentCtx",
			fields: fields{
				ws: ws,
			},
			args: args{
				wg: &sync.WaitGroup{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {

				// Allow the test to Start using a nil context and cancel function
				var ctxWithCancel context.Context
				var cancel context.CancelFunc
				if tt.args.parentCtx != nil {
					ctxWithCancel, cancel = context.WithCancel(tt.args.parentCtx)
				}

				if err := tt.fields.ws.Start(ctxWithCancel, cancel, tt.args.wg); (err != nil) != tt.wantErr {
					t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

// TestWebservice_GracefulShutdown tests the Webservice is able to gracefully shut down.
func TestWebservice_GracefulShutdown(t *testing.T) {

	ctx := context.Background()

	// Logger
	log := logger.Default()

	// Config
	cfg, err := config.New()
	if err != nil {
		t.Fatalf(`failed to create config: %v`, err)
	}

	// DB Service (memory
	db, err := memory.New(log)

	// Usecases Layer
	uc := standard.New(ctx, log, cfg, db)

	ws, err := New(
		log, &config.Config{
			APIServerPort: 5002,
		}, uc,
	)
	if err != nil {
		t.Fatalf(`failed to create webservice: %v`, err)
	}

	// Create the context with a short timeout. this will trigger the graceful shutdown of the service
	ctx, timeout := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer timeout()

	// Get a context cancel function to allow the webservice to terminate the application
	ctx, cancel := context.WithCancel(ctx)

	// Create a waitGroup to wait for the service to finish gracefully
	var wg sync.WaitGroup
	err = ws.Start(ctx, cancel, &wg)
	if err != nil {
		t.Fatalf(`failed to start webservice: %v`, err)
	}
	wg.Wait()

}
