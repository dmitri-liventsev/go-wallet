package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var users = []uint64{1, 2, 3}

type Balance struct {
	UserID uint64
	Amount float64
}

var serverList []string

func getRandomServer() string {
	return serverList[rand.Intn(len(serverList))]
}

type Job struct {
	UserID uint64
	Amount float64
}

func worker(wg *sync.WaitGroup, jobs <-chan Job, results chan<- error) {
	defer wg.Done()

	for job := range jobs {
		txID := uuid.New().String()

		statusCode, err := createTx(txID, job.UserID, job.Amount, getRandomServer())
		if statusCode != 200 {
			fmt.Println("Error creating transaction. Status code:"+strconv.Itoa(statusCode), err)
		}

		results <- err
	}
}

// waitForProcessing polls until no transactions remain in 'new' or 'locked' state,
// or until the deadline is reached.
func waitForProcessing(db *sql.DB) {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		var pending int64
		if err := db.QueryRow("SELECT COUNT(*) FROM transactions WHERE status IN ('new', 'locked')").Scan(&pending); err != nil {
			log.Printf("poll error: %v", err)
		} else if pending == 0 {
			return
		} else {
			fmt.Printf("waiting: %d transactions still pending...\n", pending)
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Println("timeout: some transactions may still be unprocessed")
}

// initUserBalance ensures a balance row exists for the user.
// Returns the current balance value: the existing one if found, or the newly inserted initial value.
func initUserBalance(db *sql.DB, userID uint64) int64 {
	const startingBalance = int64(100000)

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("tx begin error: %v", err)
	}

	var current int64
	err = tx.QueryRow("SELECT value FROM balances WHERE user_id = $1", userID).Scan(&current)
	if err == sql.ErrNoRows {
		current = startingBalance
		_, err = tx.Exec(
			"INSERT INTO balances (id, user_id, value) VALUES ($1, $2, $3)",
			uuid.New(), userID, current,
		)
		if err != nil {
			tx.Rollback()
			log.Fatalf("insert balance error: %v", err)
		}
	} else if err != nil {
		tx.Rollback()
		log.Fatalf("select balance error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("tx commit error: %v", err)
	}

	return current
}

func main() {
	numOfTransactions := flag.Int("numOfTransactions", 1000, "number of transactions to send")
	numWorkers := flag.Int("numWorkers", 20, "number of concurrent workers")
	connStr := flag.String("connStr", "user=postgres password=password dbname=txdb host=localhost port=5432 sslmode=disable", "PostgreSQL connection string")
	servers := flag.String("servers", "localhost:8081,localhost:8082", "comma-separated list of server addresses")
	flag.Parse()

	serverList = strings.Split(*servers, ",")

	rand.Seed(time.Now().UnixNano())
	runStart := time.Now()

	db, err := sql.Open("pgx", *connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// -------------------------
	// init balances for 3 users
	// -------------------------
	expected := make(map[uint64]float64, len(users))
	for _, userID := range users {
		initial := initUserBalance(db, userID)
		expected[userID] = float64(initial)
	}

	// -------------------------
	// workers
	// -------------------------
	jobs := make(chan Job, *numOfTransactions)
	results := make(chan error, *numOfTransactions)

	var wg sync.WaitGroup

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, jobs, results)
	}

	for i := 0; i < *numOfTransactions; i++ {
		userID := users[rand.Intn(len(users))]
		amount := float64(rand.Intn(2001)-1000) / 100

		jobs <- Job{
			UserID: userID,
			Amount: amount,
		}
	}

	close(jobs)

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			log.Printf("tx error: %v", err)
		}
	}

	waitForProcessing(db)

	// -------------------------
	// validate results
	// -------------------------
	ok := true
	for _, userID := range users {
		initial := int64(expected[userID])

		var finalBalance int64
		if err := db.QueryRow("SELECT value FROM balances WHERE user_id=$1", userID).Scan(&finalBalance); err != nil {
			log.Fatalf("failed to get balance for user %d: %v", userID, err)
		}

		var doneSum int64
		if err := db.QueryRow(
			"SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id=$1 AND status='done' AND created_at >= $2",
			userID, runStart,
		).Scan(&doneSum); err != nil {
			log.Fatalf("failed to get done sum for user %d: %v", userID, err)
		}

		var cancelledCount int64
		db.QueryRow(
			"SELECT COUNT(*) FROM transactions WHERE user_id=$1 AND status='cancelled' AND created_at >= $2",
			userID, runStart,
		).Scan(&cancelledCount)

		computedBalance := initial + doneSum
		status := "OK"
		if finalBalance != computedBalance {
			status = "MISMATCH"
			ok = false
		}

		fmt.Printf("User %d: balance=%.2f | computed=%.2f | cancelled=%d | %s\n",
			userID,
			float64(finalBalance)/100,
			float64(computedBalance)/100,
			cancelledCount,
			status,
		)
	}

	if ok {
		fmt.Println("All balances match.")
	} else {
		fmt.Println("BALANCE MISMATCH DETECTED.")
	}
}

func createTx(txID string, userID uint64, amount float64, host string) (int, error) {
	state := "win"
	if amount < 0 {
		state = "lose"
	}

	payload := map[string]string{
		"state":         state,
		"amount":        fmt.Sprintf("%.2f", math.Abs(amount)),
		"userID":        strconv.FormatUint(userID, 10),
		"transactionId": txID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://%s/user/%d/transaction", host, userID),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
