package main

import (
	"context"
	"flag"
	"fmt"
	"gorm.io/gorm"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"wallet/config"
	"wallet/transaction/interfaces"
	"wallet/transaction/interfaces/http"
	"wallet/transaction/internal/infrastructure/db"
	"wallet/transaction/workers"

	"goa.design/clue/debug"
	"goa.design/clue/log"
	"wallet/gen/transaction"
)

func main() {
	var dbgF = flag.Bool("debug", false, "Log request and response bodies")
	flag.Parse()

	// Setup logger. Replace logger with your own log package of choice.
	format := log.FormatJSON
	if log.IsTerminal() {
		format = log.FormatTerminal
	}
	ctx := log.Context(context.Background(), log.WithFormat(format))
	if *dbgF {
		ctx = log.Context(ctx, log.WithDebug())
		log.Debugf(ctx, "debug logs enabled")
	}

	config.Load()

	//Initialize db
	var gormdb *gorm.DB
	{
		dbConnection := db.NewConnection()
		gormdb = dbConnection.Connect(ctx)
		err := db.RunAutoMigrations(gormdb)
		if err != nil {
			panic(err)
		}
	}

	// Initialize the services.
	var txSvc transaction.Service
	{
		txSvc = interfaces.NewTxController(gormdb)
	}

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	var txEndpoints *transaction.Endpoints
	{
		txEndpoints = transaction.NewEndpoints(txSvc)
		txEndpoints.Use(debug.LogPayloads())
		txEndpoints.Use(log.Endpoint)
	}

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)

	{
		workers.RunBalanceWorker(ctx, gormdb)
		workers.RunCorrectionWorker(ctx, gormdb)
	}

	{
		addr := "http://0.0.0.0:8080/transaction"
		u, err := url.Parse(addr)
		if err != nil {
			log.Fatalf(ctx, err, "invalid URL %#v\n", addr)
		}
		http.HandleHTTPServer(ctx, u, txEndpoints, &wg, errc, *dbgF)
	}

	// Wait for signal.
	log.Printf(ctx, "exiting (%v)", <-errc)

	// Send cancellation signal to the goroutines.
	cancel()

	wg.Wait()
	log.Printf(ctx, "exited")
}
