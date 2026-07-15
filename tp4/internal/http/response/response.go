package response

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Meta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type Envelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func JSONWithMeta(w http.ResponseWriter, status int, v any, meta Meta) {
	JSON(w, status, Envelope{Data: v, Meta: &meta})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorBody{Error: code, Message: message})
}
