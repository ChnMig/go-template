package log

import (
	"go.uber.org/zap"
)

// log field
type LogField struct {
	Level  string `json:"level"`
	TS     string `json:"ts"`
	Caller string `json:"caller"`
	Msg    string `json:"msg"`
	Error  string `json:"error"`
	CustomField
}

// CustomKeyType
type CustomField struct {
	ResuestIP string `json:"request_ip"` // requested IP
	// Additional supplements based on actual needs
}

// Used as a key
// The value of the field is the value of each jsontag of CustomField
var customField = CustomField{
	ResuestIP: "request_ip",
}

// Add more according to actual needs
// Each custom field should have a function that the consumer can call
func NewRequestIPField(ip string) zap.Field {
	return zap.String(customField.ResuestIP, ip)
}
