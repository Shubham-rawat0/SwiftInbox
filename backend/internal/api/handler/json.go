package handler

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, code int, err error) {
	WriteJSON(w, code, map[string]string{
		"error": err.Error(),
	})
}