import { useState } from "react";
import { loadStripe } from "@stripe/stripe-js";
import { Elements, CardElement, useStripe, useElements } from "@stripe/react-stripe-js";
import { createPaymentIntent } from "../api/client";
import Layout from "../components/Layout";

const stripePromise = loadStripe(import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY ?? "pk_test_placeholder");

function CheckoutForm() {
  const stripe = useStripe();
  const elements = useElements();
  const [amount, setAmount] = useState(1000); // cents
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [message, setMessage] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stripe || !elements) return;

    setStatus("loading");
    setMessage("");

    try {
      const { client_secret } = await createPaymentIntent(amount);
      const card = elements.getElement(CardElement);
      if (!card) throw new Error("card element not found");

      const result = await stripe.confirmCardPayment(client_secret, {
        payment_method: { card },
      });

      if (result.error) {
        setStatus("error");
        setMessage(result.error.message ?? "Payment failed");
      } else {
        setStatus("success");
        setMessage("Payment succeeded!");
      }
    } catch (e: unknown) {
      setStatus("error");
      setMessage((e as Error).message);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Amount (USD)
        </label>
        <div className="relative">
          <span className="absolute left-3 top-2.5 text-gray-400 text-sm">$</span>
          <input
            type="number"
            min={1}
            step={1}
            value={amount / 100}
            onChange={(e) => setAmount(Math.round(Number(e.target.value) * 100))}
            className="pl-7 w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Card details
        </label>
        <div className="border border-gray-300 rounded-lg p-3">
          <CardElement options={{ style: { base: { fontSize: "14px" } } }} />
        </div>
      </div>

      {message && (
        <p className={`text-sm ${status === "success" ? "text-green-600" : "text-red-500"}`}>
          {message}
        </p>
      )}

      <button
        type="submit"
        disabled={!stripe || status === "loading"}
        className="w-full bg-indigo-600 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
      >
        {status === "loading" ? "Processing…" : `Pay $${(amount / 100).toFixed(2)}`}
      </button>
    </form>
  );
}

export default function Payment() {
  return (
    <Layout>
      <div className="max-w-md mx-auto space-y-6">
        <h1 className="text-2xl font-bold text-gray-900">Payment Demo</h1>
        <p className="text-sm text-gray-500">
          Test card: <code className="bg-gray-100 px-1 rounded">4242 4242 4242 4242</code>, any future date, any CVC.
        </p>

        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <Elements stripe={stripePromise}>
            <CheckoutForm />
          </Elements>
        </div>
      </div>
    </Layout>
  );
}
