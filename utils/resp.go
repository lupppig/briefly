package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

func JSONResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.WriteHeader(code)
	resp := Response{
		StatusCode: code,
		Message:    message,
		Data:       data,
	}

	json.NewEncoder(w).Encode(resp)
}

func InternalServerResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	resp := Response{
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
	}
	json.NewEncoder(w).Encode(resp)
}

func FerrorResponse(w http.ResponseWriter, code int, message string, err string) {
	w.WriteHeader(code)
	resp := Response{
		StatusCode: code,
		Message:    message,
		Error:      err,
	}

	json.NewEncoder(w).Encode(resp)
}
