package model

import "encoding/xml"

type ListBucketResult struct {
	XMLName  xml.Name `xml:"ListBucketResult"`
	Name     string   `xml:"Name"`
	Prefix   string   `xml:"Prefix"`
	Contents []Object `xml:"Contents"`
}

type Object struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
}

type Error struct {
	XMLName  xml.Name `xml:"Error"`
	Resource string   `xml:"Resource"`
	Code     string   `xml:"Code"`
	Message  string   `xml:"Message"`
}
