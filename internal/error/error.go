package gosssError

import (
	"encoding/xml"
	"net/http"
	"strconv"
)

// Error struct that represents the XML error response
type GosssError struct {
	XMLName  xml.Name `xml:"Error"`
	Code     string   `xml:"Code"`
	Message  string   `xml:"Message"`
	Resource string   `xml:"Resource"`
}

// SendGossError writes an error response in XML format
func SendGossError(w http.ResponseWriter, code uint, message, resource string) {
	// Create the error object
	err := GosssError{
		Code:     strconv.Itoa(int(code)),
		Message:  message,
		Resource: resource,
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(int(code)) // Use the provided status code

	// Marshal the error object to XML and write it to the response
	if errXML, err := xml.MarshalIndent(err, "", "  "); err == nil {
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"))
		w.Write(errXML) // Write the XML response
	} else {
		http.Error(w, "Failed to generate error response", http.StatusInternalServerError)
	}
}
