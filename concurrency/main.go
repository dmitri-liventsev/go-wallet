package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

const numOfTransactions = 1000
const numWorkers = 20

type Balance struct {
	ID     string
	Amount float64
}

func getRandomServer() string {
	//list of available servers
	servers := []string{"localhost:8081", "localhost:8082"}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(servers))
	return servers[index]
}

func worker(wg *sync.WaitGroup, jobs <-chan float64, results chan<- error) {
	defer wg.Done()
	for amount := range jobs {
		txID := uuid.New().String()
		statusCode, err := createTx(txID, amount, getRandomServer())
		if statusCode != 202 {
			fmt.Println("Error creating transaction. Status code:"+strconv.Itoa(statusCode), err)
		}
		results <- err
	}
}

func main() {
	connStr := "user=postgres password=password dbname=txdb host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	initialID := "0f31adad-bfb6-41d1-aeff-c110ca13cbfa"

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	var initialBalance float64
	err = tx.QueryRow("SELECT value FROM balances WHERE id = $1", initialID).Scan(&initialBalance)

	if err == sql.ErrNoRows {
		_, err = tx.Exec(`
			INSERT INTO balances (id, value)
			VALUES ($1, $2)
		`, initialID, 100000)
		initialBalance = 100000
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to insert initial balance: %v", err)
		}
	} else if err != nil {
		tx.Rollback()
		log.Fatalf("Failed to query current balance: %v", err)
	} else {
		fmt.Println("Balance already exists, no action taken.")
	}
	err = tx.Commit()

	jobs := make(chan float64, numOfTransactions)
	results := make(chan error, numOfTransactions)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, jobs, results)
	}

	rand.Seed(time.Now().UnixNano())
	var sum float64
	for i := 0; i < numOfTransactions; i++ {
		intNum := rand.Intn(2001) - 1000
		floatNum := float64(intNum) / 100
		sum += floatNum
		jobs <- floatNum
	}

	floatInitialBalance := float64(initialBalance / 100)
	expectedBalance := floatInitialBalance + sum
	fmt.Printf("Expected Balance: %.2f\n", expectedBalance*100)

	close(jobs)

	wg.Wait()

	close(results)
	for err := range results {
		if err != nil {
			log.Printf("Error occurred during transaction processing: %v", err)
		}
	}

	time.Sleep(5 * time.Second)

	var finalBalance float64
	err = db.QueryRow("SELECT value FROM balances WHERE id=$1", initialID).Scan(&finalBalance)
	if err != nil {
		log.Fatalf("Failed to retrieve final balance: %v", err)
	}

	fmt.Printf("Final Balance in Database: %.2f\n", finalBalance)
}

func createTx(txID string, amount float64, host string) (int, error) {
	action := "win"
	if amount < 0 {
		action = "lost"
	}

	payload := map[string]string{
		"state":         action,
		"amount":        fmt.Sprintf("%.2f", amount),
		"transactionId": txID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", "http://"+host+"/transaction", bytes.NewBuffer(jsonPayload))
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
