package main

import (
	"bytes"
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
)

var users = []uint64{1, 2, 3}

var serverList []string

func getRandomServer() string {
	return serverList[rand.Intn(len(serverList))]
}

type Job struct {
	UserID uint64
	Amount float64
}

func getBalance(userID uint64, server string) (float64, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/user/%d/balance", server, userID))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var body struct {
		Balance string `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}

	bal, err := strconv.ParseFloat(body.Balance, 64)
	if err != nil {
		return 0, fmt.Errorf("parse balance %q: %w", body.Balance, err)
	}
	return bal, nil
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

func main() {
	numOfTransactions := flag.Int("numOfTransactions", 1000, "number of transactions to send")
	numWorkers := flag.Int("numWorkers", 20, "number of concurrent workers")
	servers := flag.String("servers", "localhost:8081,localhost:8082", "comma-separated list of server addresses")
	flag.Parse()

	waitSeconds := *numOfTransactions / 100

	serverList = strings.Split(*servers, ",")

	rand.Seed(time.Now().UnixNano())

	// -------------------------
	// fetch initial balances
	// -------------------------
	initial := make(map[uint64]float64, len(users))
	for _, userID := range users {
		bal, err := getBalance(userID, getRandomServer())
		if err != nil {
			log.Fatalf("failed to get initial balance for user %d: %v", userID, err)
		}
		initial[userID] = bal
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

	netSent := make(map[uint64]float64, len(users))
	for i := 0; i < *numOfTransactions; i++ {
		userID := users[rand.Intn(len(users))]
		amount := float64(rand.Intn(2001)-1000) / 100
		netSent[userID] += amount

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

	// -------------------------
	// wait for async processing
	// -------------------------
	fmt.Printf("waiting %d seconds for transaction processing...\n", waitSeconds)
	time.Sleep(time.Duration(waitSeconds) * time.Second)

	// -------------------------
	// validate results
	// -------------------------
	allOK := true
	for _, userID := range users {
		balance, err := getBalance(userID, getRandomServer())
		if err != nil {
			log.Fatalf("failed to get final balance for user %d: %v", userID, err)
		}
		computed := initial[userID] + netSent[userID]
		status := "OK"
		if math.Abs(computed-balance) >= 0.01 {
			status = "MISMATCH"
			allOK = false
		}
		fmt.Printf("User %d: balance=%.2f | computed=%.2f | %s\n", userID, balance, computed, status)
	}
	if allOK {
		fmt.Println("All balances match.")
	} else {
		fmt.Println("ERRORS DETECTED: some balances do not match.")
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
