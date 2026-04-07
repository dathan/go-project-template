package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/webhook"

	"github.com/dathan/go-project-template/internal/db"
	"github.com/dathan/go-project-template/internal/db/models"
	"github.com/dathan/go-project-template/internal/server/middleware"
)

// PaymentHandler handles Stripe payment intents and webhooks.
type PaymentHandler struct {
	store         db.Store
	stripeKey     string
	webhookSecret string
}

func NewPaymentHandler(store db.Store, stripeKey, webhookSecret string) *PaymentHandler {
	stripe.Key = stripeKey
	return &PaymentHandler{store: store, stripeKey: stripeKey, webhookSecret: webhookSecret}
}

// CreateIntent creates a Stripe PaymentIntent and records it in the DB.
// POST /api/v1/payments/intent
func (h *PaymentHandler) CreateIntent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Amount   int64  `json:"amount"`   // cents
		Currency string `json:"currency"` // default "usd"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if req.Currency == "" {
		req.Currency = "usd"
	}

	user := middleware.UserFromContext(r.Context())

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create payment intent")
		return
	}

	payment := &models.Payment{
		ID:              uuid.New(),
		UserID:          user.ID,
		StripePaymentID: pi.ID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          models.PaymentStatusPending,
	}
	if err := h.store.Payments().Create(r.Context(), payment); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record payment")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"client_secret": pi.ClientSecret,
		"payment_id":    payment.ID,
	})
}

// Webhook processes Stripe webhook events (payment_intent.succeeded, etc.).
// POST /api/v1/webhooks/stripe
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	const maxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "error reading body")
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.webhookSecret)
	if err != nil {
		writeError(w, http.StatusBadRequest, "webhook signature verification failed")
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse payment intent")
			return
		}
		if err := h.store.Payments().UpdateStatus(r.Context(), pi.ID, models.PaymentStatusSucceeded); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update payment status")
			return
		}
		// Mark the user as paid
		payment, err := h.store.Payments().GetByStripeID(r.Context(), pi.ID)
		if err == nil {
			now := time.Now()
			_ = h.store.Users().SetPaidAt(r.Context(), payment.UserID, now)
		}

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err == nil {
			_ = h.store.Payments().UpdateStatus(r.Context(), pi.ID, models.PaymentStatusFailed)
		}
	}

	w.WriteHeader(http.StatusOK)
}
