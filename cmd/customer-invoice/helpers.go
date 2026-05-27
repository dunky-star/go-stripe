package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

// writeJSON writes arbitrary data out as JSON.
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(data); err != nil {
		return err
	}

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	const maxBytes = 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(data); err != nil {
		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

// badRequest sends a JSON response with status http.StatusBadRequest, describing the error.
func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, http.StatusBadRequest, payload)
}

func (app *application) CreateDirIfNotExist(path string) error {
	const mode = 0755

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}

	return nil
}
