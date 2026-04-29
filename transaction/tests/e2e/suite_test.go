package e2e_test

import (
	"context"
	"net"
	"net/url"
	"sync"
	"testing"
	"wallet/config"
	"wallet/gen/wallet"
	"wallet/transaction/interfaces"
	"wallet/transaction/interfaces/http"
	"wallet/transaction/internal/infrastructure/db"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"goa.design/clue/log"
	"gorm.io/gorm"
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
	txEndpoints := wallet.NewEndpoints(txSvc)
	u, err := url.Parse(addr)
	if err != nil {
		panic("failed to parse address")
	}
	var wg sync.WaitGroup
	errc := make(chan error)

	http.HandleHTTPServer(ctx, u, txEndpoints, &wg, errc, false)

	host := u.Host
	Eventually(func() error {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			GinkgoWriter.Printf("waiting for server on %s: %v\n", host, err)
			return err
		}
		conn.Close()
		return nil
	}, "5s", "10ms").Should(Succeed())

	DeferCleanup(func() {
		cancel()
		wg.Wait()
	})
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "wallet/tests")
}
