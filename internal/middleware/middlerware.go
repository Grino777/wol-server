// Package middleware defines common http middlewares.
package middleware

import (
	"net/http"
)

// contextKey is a unique type to avoid clashing with other packages that use
// context's to pass data.
type contextKey string

// Middleware is a function which receives an http.Handler and returns another http.Handler.
// Typically, the returned handler is a closure which does something with the http.ResponseWriter and http.Request passed
// to it, and then calls the handler passed as parameter to the MiddlewareFunc.
type Middleware = func(http.Handler) http.Handler
