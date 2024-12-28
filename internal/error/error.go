package gosssError

import (
	"encoding/xml"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type ErrorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource"`
	RequestID string   `xml:"RequestId"`
	TimeStamp string   `xml:"Timestamp"`
}

type ErrorLogger struct {
	logger *log.Logger
}

func NewErrorLogger(l *log.Logger) *ErrorLogger {
	return &ErrorLogger{logger: l}
}

func (el *ErrorLogger) LogError(err error, code uint, resource string) {
	_, file, line, _ := runtime.Caller(1)
	el.logger.Printf("[ERROR] %s:%d - Code: %d, Resource: %s, Error: %v",
		file, line, code, resource, err)
}

func SendGossError(w http.ResponseWriter, code uint, message, resource string) {
	err := ErrorResponse{
		Code:      strconv.Itoa(int(code)),
		Message:   message,
		Resource:  resource,
		RequestID: generateRequestID(),
		TimeStamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(int(code))
	log.Printf("Sending error response: %v", err)
	if errXML, err := xml.MarshalIndent(err, "", "  "); err == nil {
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"))
		w.Write(errXML)
	} else {
		log.Printf("Failed to generate error response: %v", err)
		http.Error(w, "Failed to generate error response", http.StatusInternalServerError)
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
