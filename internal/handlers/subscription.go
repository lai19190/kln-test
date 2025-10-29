package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"kln-test/internal/config"
	"kln-test/internal/worker"

	"github.com/go-playground/validator/v10"
)

// SubscriptionRequest represents the subscription payload
type SubscriptionRequest struct {
	ConsumerID  string   `json:"consumerId" validate:"required"`
	Topics      []string `json:"topics" validate:"required,min=1,dive,required"`
	DeliveryURL string   `json:"deliveryUrl" validate:"required,url"`
}

// SubscriptionResponse represents the subscription response
type SubscriptionResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

// SubscriptionHandler handles shipping event subscriptions
type SubscriptionHandler struct {
	validator *validator.Validate
	pool      *worker.Pool[SubscriptionRequest]
	cfg       *config.Config
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(cfg *config.Config) *SubscriptionHandler {
	return &SubscriptionHandler{
		validator: validator.New(),
		pool:      worker.NewPool[SubscriptionRequest](cfg),
		cfg:       cfg,
	}
}

// ServeHTTP handles HTTP requests for subscriptions
func (h *SubscriptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Create a job for async processing
	job := worker.Job[SubscriptionRequest]{
		ID:      req.ConsumerID,
		Payload: req,
		Process: h.processSubscription,
	}

	if err := h.pool.Submit(job); err != nil {
		http.Error(w, "Server is busy, try again later", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(SubscriptionResponse{
		Message: "Subscription request accepted",
		ID:      req.ConsumerID,
	})
}

// processSubscription handles the subscription processing
func (h *SubscriptionHandler) processSubscription(ctx context.Context, payload SubscriptionRequest) error {
	// push the subscription to external service, e.g. cache DB, message queue, data lake, etc.
	time.Sleep(5 * time.Second)

	// _, err := svc.SendMessage(&sqs.SendMessageInput{
	// 	MessageAttributes: map[string]*sqs.MessageAttributeValue{
	// 		"ConsumerID": &sqs.MessageAttributeValue{
	// 			DataType:    aws.String("String"),
	// 			StringValue: aws.String(payload.ConsumerID),
	// 		},
	// 		...
	// 	},
	// 	MessageBody: aws.String("Shipping event subscription"),
	// 	QueueUrl:    queueURL,
	// })
	return nil
}
