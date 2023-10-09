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
	ClientIP string `json:"client_ip"`
	// Additional supplements based on actual needs
}

// Used as a key
// The value of the field is the value of each jsontag of CustomField
var customField = CustomField{
	ClientIP: "client_ip",
}

// Add more according to actual needs
// Each custom field should have a function that the consumer can call
func NewClientIPField(ip string) zap.Field {
	return zap.String(customField.ClientIP, ip)
}
