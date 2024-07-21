package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"wallet/transaction/internal/domain/entities"
)

func createTx(txID string, amount float64, action string) int {
	GinkgoHelper()

	payload := map[string]string{
		"state":         action,
		"amount":        fmt.Sprintf("%.2f", amount),
		"transactionId": txID,
	}

	jsonPayload, err := json.Marshal(payload)
	Expect(err).NotTo(HaveOccurred())

	req, err := http.NewRequest("POST", "http://localhost:8080/transaction", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return http.StatusInternalServerError
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", entities.Game)

	client := &http.Client{}
	resp, err := client.Do(req)
	Expect(err).NotTo(HaveOccurred())

	defer resp.Body.Close()

	return resp.StatusCode
}
