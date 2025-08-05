package integrations

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/logger"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type PhonePeClient struct {
	httpClient *http.Client
}

func NewPhonePeClient() *PhonePeClient {
	return &PhonePeClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InitiatePayment initiates payment with PhonePe
func (p *PhonePeClient) InitiatePayment(config *models.PaymentGatewayConfig, amount float64, currency, transactionID string, userID int64) (map[string]interface{}, error) {
	baseURL := "https://api-preprod.phonepe.com/apis/pg-sandbox/pg/v1/pay"
	if config.IsLive {
		baseURL = "https://api.phonepe.com/apis/hermes/pg/v1/pay"
		logger.Info("Using PhonePe production environment")
	} else {
		logger.Info("Using PhonePe test environment")
	}

	// Convert amount to paisa (smallest currency unit)
	amountInPaisa := int64(amount * 100)

	// Create merchant transaction ID (must be unique)
	merchantTransactionID := fmt.Sprintf("MT_%s_%d", transactionID, time.Now().Unix())

	// Prepare payment payload
	paymentPayload := map[string]interface{}{
		"merchantId":            config.Key1, // client_id
		"merchantTransactionId": merchantTransactionID,
		"merchantUserId":        fmt.Sprintf("USER_%d", userID),
		"amount":                amountInPaisa,
		"redirectUrl":           "https://webhook.site/redirect-url", // Replace with your redirect URL
		"redirectMode":          "POST",
		"callbackUrl":           "https://webhook.site/callback-url", // Replace with your callback URL
		"mobileNumber":          "", // Will be filled by user in PhonePe app
		"paymentInstrument": map[string]interface{}{
			"type": "PAY_PAGE",
		},
	}

	// Convert payload to JSON string
	payloadJSON, err := json.Marshal(paymentPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment payload: %v", err)
	}

	// Base64 encode the payload
	base64Payload := base64.StdEncoding.EncodeToString(payloadJSON)

	// Create checksum
	checksumString := base64Payload + "/pg/v1/pay" + config.Key2 // client_secret
	checksum := p.generateChecksum(checksumString)

	// Prepare request body
	requestBody := map[string]interface{}{
		"request": base64Payload,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-VERIFY", checksum)
	req.Header.Set("X-MERCHANT-ID", config.Key1)

	// Make request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("PhonePe payment initiation failed", "status", resp.StatusCode, "response", string(body))
		return nil, fmt.Errorf("phonepe API error: %s", string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Check if success
	success, ok := response["success"].(bool)
	if !ok || !success {
		return nil, fmt.Errorf("payment initiation failed: %v", response["message"])
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data")
	}

	// Return data needed for Flutter app
	paymentData := map[string]interface{}{
		"merchant_transaction_id": merchantTransactionID,
		"payment_url":            data["instrumentResponse"].(map[string]interface{})["redirectInfo"].(map[string]interface{})["url"],
		"merchant_id":            config.Key1,
		"transaction_id":         transactionID,
		"amount":                 amountInPaisa,
		"currency":               currency,
	}

	logger.Info("PhonePe payment initiated successfully", "merchant_transaction_id", merchantTransactionID, "amount", amount)

	return paymentData, nil
}

// VerifyPayment verifies payment with PhonePe
func (p *PhonePeClient) VerifyPayment(config *models.PaymentGatewayConfig, gatewayData map[string]interface{}) (bool, string, map[string]interface{}, error) {
	merchantTransactionID, ok := gatewayData["merchant_transaction_id"].(string)
	if !ok {
		return false, "", nil, fmt.Errorf("missing merchant_transaction_id")
	}

	// Check payment status
	status, paymentDetails, err := p.CheckPaymentStatus(config, merchantTransactionID)
	if err != nil {
		return false, merchantTransactionID, gatewayData, fmt.Errorf("failed to check payment status: %v", err)
	}

	success := status == "PAYMENT_SUCCESS"

	logger.Info("PhonePe payment verification completed", "merchant_transaction_id", merchantTransactionID, "status", status, "success", success)

	return success, merchantTransactionID, paymentDetails, nil
}

// CheckPaymentStatus checks payment status with PhonePe
func (p *PhonePeClient) CheckPaymentStatus(config *models.PaymentGatewayConfig, merchantTransactionID string) (string, map[string]interface{}, error) {
	baseURL := "https://api-preprod.phonepe.com/apis/pg-sandbox/pg/v1/status"
	if config.IsLive {
		baseURL = "https://api.phonepe.com/apis/hermes/pg/v1/status"
	}

	// Create checksum for status check
	checksumString := fmt.Sprintf("/pg/v1/status/%s/%s%s", config.Key1, merchantTransactionID, config.Key2)
	checksum := p.generateChecksum(checksumString)

	// Create request URL
	url := fmt.Sprintf("%s/%s/%s", baseURL, config.Key1, merchantTransactionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-VERIFY", checksum)
	req.Header.Set("X-MERCHANT-ID", config.Key1)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("PhonePe status check failed", "status", resp.StatusCode, "response", string(body))
		return "", nil, fmt.Errorf("phonepe API error: %s", string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %v", err)
	}

	success, ok := response["success"].(bool)
	if !ok || !success {
		return "FAILED", response, nil
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return "FAILED", response, fmt.Errorf("invalid response data")
	}

	state, ok := data["state"].(string)
	if !ok {
		return "FAILED", response, fmt.Errorf("invalid payment state")
	}

	return state, response, nil
}

// generateChecksum generates SHA256 checksum for PhonePe API
func (p *PhonePeClient) generateChecksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]) + "###1" // ###1 is the salt key index
}

// HandleCallback handles PhonePe callback/webhook
func (p *PhonePeClient) HandleCallback(config *models.PaymentGatewayConfig, payload []byte, checksum string) (map[string]interface{}, error) {
	// Parse the callback payload
	var callbackData map[string]interface{}
	if err := json.Unmarshal(payload, &callbackData); err != nil {
		return nil, fmt.Errorf("failed to parse callback payload: %v", err)
	}

	// Verify checksum
	response, ok := callbackData["response"].(string)
	if !ok {
		return nil, fmt.Errorf("missing response in callback")
	}

	// Decode base64 response
	decodedResponse, err := base64.StdEncoding.DecodeString(response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(decodedResponse, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse decoded response: %v", err)
	}

	// Verify checksum
	checksumString := response + "/pg/v1/status" + config.Key2
	expectedChecksum := p.generateChecksum(checksumString)

	if checksum != expectedChecksum {
		return nil, fmt.Errorf("checksum verification failed")
	}

	logger.Info("PhonePe callback verified successfully", "merchant_transaction_id", responseData["merchantTransactionId"])

	return responseData, nil
}

// RefundPayment creates a refund for a payment
func (p *PhonePeClient) RefundPayment(config *models.PaymentGatewayConfig, merchantTransactionID string, amount float64, reason string) (map[string]interface{}, error) {
	baseURL := "https://api-preprod.phonepe.com/apis/pg-sandbox/pg/v1/refund"
	if config.IsLive {
		baseURL = "https://api.phonepe.com/apis/hermes/pg/v1/refund"
	}

	amountInPaisa := int64(amount * 100)
	refundTransactionID := fmt.Sprintf("RF_%s_%d", merchantTransactionID, time.Now().Unix())

	// Prepare refund payload
	refundPayload := map[string]interface{}{
		"merchantId":                   config.Key1,
		"merchantTransactionId":        refundTransactionID,
		"originalTransactionId":        merchantTransactionID,
		"amount":                       amountInPaisa,
		"callbackUrl":                  "https://webhook.site/refund-callback", // Replace with your callback URL
	}

	payloadJSON, err := json.Marshal(refundPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refund payload: %v", err)
	}

	base64Payload := base64.StdEncoding.EncodeToString(payloadJSON)
	checksumString := base64Payload + "/pg/v1/refund" + config.Key2
	checksum := p.generateChecksum(checksumString)

	requestBody := map[string]interface{}{
		"request": base64Payload,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-VERIFY", checksum)
	req.Header.Set("X-MERCHANT-ID", config.Key1)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("PhonePe refund failed", "status", resp.StatusCode, "response", string(body))
		return nil, fmt.Errorf("phonepe API error: %s", string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	success, ok := response["success"].(bool)
	if !ok || !success {
		return nil, fmt.Errorf("refund failed: %v", response["message"])
	}

	logger.Info("PhonePe refund created successfully", "refund_transaction_id", refundTransactionID, "amount", amount)

	return response, nil
}

// ValidateCredentials validates PhonePe API credentials
func (p *PhonePeClient) ValidateCredentials(clientID, clientSecret string) error {
	// PhonePe doesn't have a direct validate credentials endpoint
	// We can do a dummy status check with a non-existent transaction
	baseURL := "https://api-preprod.phonepe.com/apis/pg-sandbox/pg/v1/status"
	
	dummyTransactionID := "TEST_" + strconv.FormatInt(time.Now().Unix(), 10)
	checksumString := fmt.Sprintf("/pg/v1/status/%s/%s%s", clientID, dummyTransactionID, clientSecret)
	checksum := p.generateChecksum(checksumString)

	url := fmt.Sprintf("%s/%s/%s", baseURL, clientID, dummyTransactionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-VERIFY", checksum)
	req.Header.Set("X-MERCHANT-ID", clientID)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// If we get 401 or 403, credentials are invalid
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid credentials")
	}

	// Any other status is acceptable for credential validation
	// (404 is expected for non-existent transaction, 200 would mean valid but transaction doesn't exist)
	return nil
}