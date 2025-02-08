package tripay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mdayat/running-man/configs/env"
)

var paymentChannel = struct {
	qris string
}{
	qris: "QRIS",
}

type OrderedItem struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type NewTransactionBodyParams struct {
	MerchantRef   string
	CustomerName  string
	CustomerEmail string
	TotalAmount   int
	OrderedItems  []OrderedItem
}

type transactionBody struct {
	Method        string        `json:"method"`
	MerchantRef   string        `json:"merchant_ref"`
	Amount        int           `json:"amount"`
	CustomerName  string        `json:"customer_name"`
	CustomerEmail string        `json:"customer_email"`
	OrderItems    []OrderedItem `json:"order_items"`
	Signature     string        `json:"signature"`
}

func createSignature(merchantRef string, amount int) string {
	key := []byte(env.TripayPrivateKey)
	message := fmt.Sprintf("%s%s%d", env.TripayMerchantCode, merchantRef, amount)

	hash := hmac.New(sha256.New, key)
	hash.Write([]byte(message))

	return hex.EncodeToString(hash.Sum(nil))
}

func NewTransactionBody(arg NewTransactionBodyParams) transactionBody {
	signature := createSignature(arg.MerchantRef, arg.TotalAmount)
	return transactionBody{
		Method:        paymentChannel.qris,
		MerchantRef:   arg.MerchantRef,
		Amount:        arg.TotalAmount,
		CustomerName:  arg.CustomerName,
		CustomerEmail: arg.CustomerEmail,
		Signature:     signature,
		OrderItems:    arg.OrderedItems,
	}
}

type successfulTransaction struct {
	Reference   string `json:"reference"`
	Amount      int    `json:"amount"`
	ExpiredTime int    `json:"expired_time"`
	QrURL       string `json:"qr_url"`
}

func RequestTransaction(ctx context.Context, body transactionBody) (response successfulTransaction, _ error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return response, fmt.Errorf("failed to encode request body to request tripay transaction: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, env.TripayURL, &buf)
	if err != nil {
		return response, fmt.Errorf("failed to create new http post request to request tripay transaction: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", env.TripayAPIKey))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return response, fmt.Errorf("failed to send http post request to request tripay transaction: %w", err)
	}
	defer res.Body.Close()

	var payload struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}

	if err = json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return response, fmt.Errorf("failed to decode response body of tripay transaction request: %w", err)
	}

	if !payload.Success {
		errMsg := fmt.Sprintf("failed response of tripay transaction request: %s", payload.Message)
		return response, errors.New(errMsg)
	}

	if err = json.Unmarshal(payload.Data, &response); err != nil {
		return response, fmt.Errorf("failed to unmarshal successful response of tripay transaction request: %w", err)
	}

	return response, nil
}
