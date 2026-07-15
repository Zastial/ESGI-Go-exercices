// Package response provides a stable JSON envelope for all API responses.
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Envelope is the stable JSON shape returned by every endpoint.
type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
	Meta  *Meta      `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
	Total  int `json:"total,omitempty"`
}

// JSON writes data wrapped in the standard envelope.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, Envelope{Data: data})
}

// JSONWithMeta writes data plus pagination metadata wrapped in the standard envelope.
func JSONWithMeta(w http.ResponseWriter, status int, data any, meta Meta) {
	write(w, status, Envelope{Data: data, Meta: &meta})
}

// Error writes an error wrapped in the standard envelope.
func Error(w http.ResponseWriter, status int, code, message string) {
	write(w, status, Envelope{Error: &ErrorBody{Code: code, Message: message}})
}

func write(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(env); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
