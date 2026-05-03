package main

import "net/http"

func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// Auth requires a logged-in server session (userID), same idea as source-code chi middleware.
func (app *application) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.Session.Exists(r.Context(), "userID") {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
