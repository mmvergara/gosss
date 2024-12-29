package gosssError

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Resource  string `json:"resource"`
	TimeStamp string `json:"timestamp"`
}

type ErrorLogger struct {
	logger *log.Logger
}

func NewErrorLogger(l *log.Logger) *ErrorLogger {
	return &ErrorLogger{logger: l}
}

func SendGossError(w http.ResponseWriter, code uint, message, resource string) {
	errorResponse := ErrorResponse{
		Code:      strconv.Itoa(int(code)),
		Message:   message,
		Resource:  resource,
		TimeStamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(code))
	log.Printf("Sending error response: %+v", errorResponse)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Failed to generate error response: %v", err)
		http.Error(w, "Failed to generate error response", http.StatusInternalServerError)
		return
	}
}
