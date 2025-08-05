package integrations

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/logger"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RazorpayClient struct {
	httpClient *http.Client
}

func NewRazorpayClient() *RazorpayClient {
	return &RazorpayClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateOrder creates a payment order with Razorpay
func (r *RazorpayClient) CreateOrder(config *models.PaymentGatewayConfig, amount float64, currency, receiptID string) (map[string]interface{}, error) {
	baseURL := "https://api.razorpay.com/v1"
	if !config.IsLive {
		// Test environment uses same URL but with test keys
		logger.Info("Using Razorpay test environment")
	}

	// Convert amount to paisa (smallest currency unit)
	amountInPaisa := int64(amount * 100)

	orderData := map[string]interface{}{
		"amount":   amountInPaisa,
		"currency": currency,
		"receipt":  receiptID,
		"notes": map[string]string{
			"transaction_id": receiptID,
			"purpose":        "add_money",
		},
	}

	jsonData, err := json.Marshal(orderData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order data: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL+"/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(config.Key1, config.Key2) // key_id, key_secret

	// Make request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Razorpay order creation failed", "status", resp.StatusCode, "response", string(body))
		return nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var orderResponse map[string]interface{}
	if err := json.Unmarshal(body, &orderResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Return data needed for Flutter app
	paymentData := map[string]interface{}{
		"order_id":     orderResponse["id"],
		"amount":       orderResponse["amount"],
		"currency":     orderResponse["currency"],
		"key_id":       config.Key1,
		"name":         "Fantasy Esports",
		"description":  "Add money to wallet",
		"prefill": map[string]interface{}{
			"contact": "",
			"email":   "",
		},
		"theme": map[string]interface{}{
			"color": "#3399cc",
		},
		"notes": orderResponse["notes"],
	}

	logger.Info("Razorpay order created successfully", "order_id", orderResponse["id"], "amount", amount)

	return paymentData, nil
}

// VerifyPayment verifies payment signature and fetches payment details
func (r *RazorpayClient) VerifyPayment(config *models.PaymentGatewayConfig, gatewayData map[string]interface{}) (bool, string, map[string]interface{}, error) {
	// Extract payment data
	paymentID, ok := gatewayData["razorpay_payment_id"].(string)
	if !ok {
		return false, "", nil, fmt.Errorf("missing razorpay_payment_id")
	}

	orderID, ok := gatewayData["razorpay_order_id"].(string)
	if !ok {
		return false, "", nil, fmt.Errorf("missing razorpay_order_id")
	}

	signature, ok := gatewayData["razorpay_signature"].(string)
	if !ok {
		return false, "", nil, fmt.Errorf("missing razorpay_signature")
	}

	// Verify signature
	if !r.verifySignature(orderID, paymentID, signature, config.Key2) {
		logger.Error("Razorpay signature verification failed", "payment_id", paymentID)
		return false, paymentID, gatewayData, fmt.Errorf("signature verification failed")
	}

	// Fetch payment details from Razorpay
	baseURL := "https://api.razorpay.com/v1"
	req, err := http.NewRequest("GET", baseURL+"/payments/"+paymentID, nil)
	if err != nil {
		return false, paymentID, gatewayData, fmt.Errorf("failed to create payment fetch request: %v", err)
	}

	req.SetBasicAuth(config.Key1, config.Key2)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return false, paymentID, gatewayData, fmt.Errorf("failed to fetch payment details: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, paymentID, gatewayData, fmt.Errorf("failed to read payment response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to fetch payment details", "status", resp.StatusCode, "response", string(body))
		return false, paymentID, gatewayData, fmt.Errorf("failed to fetch payment details: %s", string(body))
	}

	var paymentDetails map[string]interface{}
	if err := json.Unmarshal(body, &paymentDetails); err != nil {
		return false, paymentID, gatewayData, fmt.Errorf("failed to parse payment details: %v", err)
	}

	// Check payment status
	status, ok := paymentDetails["status"].(string)
	if !ok {
		return false, paymentID, paymentDetails, fmt.Errorf("invalid payment status")
	}

	success := status == "captured"

	logger.Info("Razorpay payment verification completed", "payment_id", paymentID, "status", status, "success", success)

	return success, paymentID, paymentDetails, nil
}

// verifySignature verifies Razorpay webhook signature
func (r *RazorpayClient) verifySignature(orderID, paymentID, signature, secret string) bool {
	// Create the string to be signed
	message := orderID + "|" + paymentID

	// Create HMAC SHA256 hash
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// WebhookHandler handles Razorpay webhooks
func (r *RazorpayClient) HandleWebhook(config *models.PaymentGatewayConfig, payload []byte, signature string) (map[string]interface{}, error) {
	// Verify webhook signature
	if !r.verifyWebhookSignature(payload, signature, config.Key2) {
		return nil, fmt.Errorf("webhook signature verification failed")
	}

	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %v", err)
	}

	// Extract event type
	event, ok := webhookData["event"].(string)
	if !ok {
		return nil, fmt.Errorf("missing event type in webhook")
	}

	logger.Info("Razorpay webhook received", "event", event)

	return webhookData, nil
}

func (r *RazorpayClient) verifyWebhookSignature(payload []byte, signature, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GetPaymentStatus fetches payment status from Razorpay
func (r *RazorpayClient) GetPaymentStatus(config *models.PaymentGatewayConfig, paymentID string) (string, map[string]interface{}, error) {
	baseURL := "https://api.razorpay.com/v1"
	
	req, err := http.NewRequest("GET", baseURL+"/payments/"+paymentID, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(config.Key1, config.Key2)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var paymentDetails map[string]interface{}
	if err := json.Unmarshal(body, &paymentDetails); err != nil {
		return "", nil, fmt.Errorf("failed to parse response: %v", err)
	}

	status, ok := paymentDetails["status"].(string)
	if !ok {
		return "", paymentDetails, fmt.Errorf("invalid payment status")
	}

	return status, paymentDetails, nil
}

// RefundPayment creates a refund for a payment
func (r *RazorpayClient) RefundPayment(config *models.PaymentGatewayConfig, paymentID string, amount float64, reason string) (map[string]interface{}, error) {
	baseURL := "https://api.razorpay.com/v1"
	amountInPaisa := int64(amount * 100)

	refundData := map[string]interface{}{
		"amount": amountInPaisa,
		"notes": map[string]string{
			"reason": reason,
		},
	}

	jsonData, err := json.Marshal(refundData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refund data: %v", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/payments/"+paymentID+"/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(config.Key1, config.Key2)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Razorpay refund failed", "status", resp.StatusCode, "response", string(body))
		return nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var refundResponse map[string]interface{}
	if err := json.Unmarshal(body, &refundResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	logger.Info("Razorpay refund created successfully", "refund_id", refundResponse["id"], "amount", amount)

	return refundResponse, nil
}

// ValidateCredentials validates Razorpay API credentials
func (r *RazorpayClient) ValidateCredentials(keyID, keySecret string) error {
	req, err := http.NewRequest("GET", "https://api.razorpay.com/v1/payments", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(keyID, keySecret)
	req.URL.RawQuery = "count=1" // Limit to 1 result

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid credentials")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s", string(body))
	}

	return nil
}