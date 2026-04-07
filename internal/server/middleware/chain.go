// Package middleware provides composable http.Handler wrappers.
// Business logic never imports this package — it stays in plain handlers.
package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order: first middleware is outermost.
//
//	Chain(h, A, B, C) → A(B(C(h)))
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
