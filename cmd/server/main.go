package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dathan/go-project-template/internal/stripe"
)

func main() {
	// It's recommended to load the API key from a secure source,
	// such as an environment variable or a secret management system.
	apiKey := os.Getenv("STRIPE_API_KEY")
	if apiKey == "" {
		log.Fatal("STRIPE_API_KEY environment variable not set")
	}

	// Create a new Stripe client.
	client := stripe.NewClient(apiKey)

	// Example: Create a Payment Intent
	amount := int64(2000) // $20.00
	currency := "usd"
	pi, err := client.CreatePaymentIntent(amount, currency)
	if err != nil {
		log.Fatalf("Error creating payment intent: %v", err)
	}
	fmt.Printf("Successfully created payment intent: %s\n", pi.ID)

	// In a real application, you would pass the client secret (`pi.ClientSecret`)
	// to your frontend to finalize the payment. After the user completes the
	// payment on the frontend, you would receive the payment intent ID
	// and could confirm its status on the backend.

	// Example: Confirm a Payment Intent (simulated)
	// In a real scenario, you'd get the payment intent ID from a webhook
	// or after the frontend confirms the payment.
	confirmedPI, err := client.ConfirmPaymentIntent(pi.ID)
	if err != nil {
		// Note: In a real-world scenario, some errors are expected,
		// such as the payment requiring further action.
		log.Fatalf("Error confirming payment intent: %v", err)
	}

	if confirmedPI.Status == "succeeded" {
		fmt.Printf("Payment intent %s successfully confirmed and succeeded!\n", confirmedPI.ID)
		// Here you would allocate the digital product to the customer.
		fmt.Println("Allocating digital product...")
	} else {
		fmt.Printf("Payment intent %s status: %s\n", confirmedPI.ID, confirmedPI.Status)
	}
}
