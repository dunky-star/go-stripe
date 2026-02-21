package main

import (
	"net/http"
	"slices"
	"strings"
)

var allowedOrigins = []string{
	"https://localhost:4000",
	"http://localhost:4000",
}

var allowedMethods = []string{
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"OPTIONS",
}

var allowedHeaders = []string{
	"Content-Type",
	"Authorization",
	"X-CSRF-Token",
}

const (
	allowCredentials = "false"
	maxAge           = "300"
)

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)
		w.Header().Set("Access-Control-Max-Age", maxAge)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isOriginAllowed(origin string) bool {
	return slices.Contains(allowedOrigins, origin)
}
