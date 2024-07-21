package e2e_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"goa.design/clue/log"
	"gorm.io/gorm"
	"net/url"
	"sync"
	"testing"
	"wallet/config"
	"wallet/gen/transaction"
	"wallet/transaction/interfaces"
	"wallet/transaction/interfaces/http"
	"wallet/transaction/internal/infrastructure/db"
)

var addr = "http://0.0.0.0:8081/transaction"
var DB *gorm.DB

var _ = BeforeSuite(func(ctx context.Context) {
	DB = connectToTestDB(ctx)
	runServer()
})

func connectToTestDB(ctx context.Context) *gorm.DB {
	config.Load()

	//Initialize db
	dbConnection := db.NewConnection()
	gormdb := dbConnection.Connect(ctx)

	DeferCleanup(func() {
		sqlDB, _ := gormdb.DB()
		_ = sqlDB.Close()
	})

	return gormdb
}

func runServer() {
	format := log.FormatJSON
	ctx := log.Context(context.Background(), log.WithFormat(format))
	ctx, cancel := context.WithCancel(ctx)

	txSvc := interfaces.NewTxController(DB)
	txEndpoints := transaction.NewEndpoints(txSvc)
	u, err := url.Parse(addr)
	if err != nil {
		panic("failed to parse address")
	}
	var wg sync.WaitGroup
	errc := make(chan error)

	http.HandleHTTPServer(ctx, u, txEndpoints, &wg, errc, false)

	DeferCleanup(func() {
		cancel()
		wg.Wait()
	})
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "wallet/tests")
}
