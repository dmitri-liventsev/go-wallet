package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
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

func getRandomServer() string {
	servers := []string{"localhost:8081", "localhost:8082"}
	return servers[rand.Intn(len(servers))]
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
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	connStr := "user=postgres password=password dbname=txdb host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("pgx", connStr)
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
		userId := users[rand.Intn(len(users))]

		intNum := rand.Intn(2001) - 1000
		floatNum := float64(intNum)
		amount := floatNum / 100

		expected[userId] += floatNum

		jobs <- Job{
			UserID: userId,
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

	time.Sleep(3 * time.Second)

	// -------------------------
	// validate results
	// -------------------------
	for _, userId := range users {
		var final int64
		err := db.QueryRow("SELECT value FROM balances WHERE user_id=$1", userId).Scan(&final)
		if err != nil {
			log.Fatalf("failed to get balance: %v", err)
		}

		fmt.Printf("User %d final balance: %.2f | expected: %.2f\n",
			userId,
			float64(final)/100,
			expected[userId]/100,
		)
	}
}

func createTx(txID string, userID uint64, amount float64, host string) (int, error) {
	state := "win"
	if amount < 0 {
		state = "lose"
	}

	payload := map[string]string{
		"state":         state,
		"amount":        fmt.Sprintf("%.2f", amount),
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
