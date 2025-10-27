// https://google-cloud.gitbook.io/api-design-guide/errors
package response

type responseData struct {
	Code      int         `json:"code"`
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Timestamp int64       `json:"timestamp"`
	Detail    interface{} `json:"detail,omitempty"`
	Total     *int        `json:"total,omitempty"`
}

// Common error messages

// success
var OK = responseData{
	Code:    200,
	Status:  "OK",
	Message: "Successful",
}

// The data sent by the client contains illegal parameters.
// View error messages and error details for more information.
var INVALID_ARGUMENT = responseData{
	Code:    400,
	Status:  "INVALID_ARGUMENT",
	Message: "Request parameter error",
}

// The current state of the system does not allow execution of the current request
// such as deleting a non-empty directory.
var FAILED_PRECONDITION = responseData{
	Code:    400,
	Status:  "FAILED_PRECONDITION",
	Message: "Unable to execute client request",
}

// The client specified an illegal scope.
var OUT_OF_RANGE = responseData{
	Code:    400,
	Status:  "OUT_OF_RANGE",
	Message: "Client access limit exceeded",
}

// The request failed authentication because of missing, invalid, or expired OAuth tokens.
var UNAUTHENTICATED = responseData{
	Code:    401,
	Status:  "UNAUTHENTICATED",
	Message: "authentication failed",
}

// The client does not have sufficient permissions.
// This could be because the OAuth token does not have the correct scope
// or the client does not have permissions, or the API is disabled for client code.
var PERMISSION_DENIED = responseData{
	Code:    403,
	Status:  "PERMISSION_DENIED",
	Message: "Insufficient client permissions",
}

// The particular resource was not found or the request was rejected for reasons that were not disclosed (e.g. whitelisting).
var NOT_FOUND = responseData{
	Code:    404,
	Status:  "NOT_FOUND",
	Message: "resource does not exist",
}

// Concurrency conflicts
// such as read-modify-write conflicts.
var ABORTED = responseData{
	Code:    409,
	Status:  "ABORTED",
	Message: "data processing conflict",
}

// The resource the client is trying to create already exists.
var ALREADY_EXISTS = responseData{
	Code:    409,
	Status:  "ALREADY_EXISTS",
	Message: "resource already exists",
}

// Resource quota is insufficient or rate limit is not reached.
var RESOURCE_EXHAUSTED = responseData{
	Code:    429,
	Status:  "RESOURCE_EXHAUSTED",
	Message: "Resource quota is insufficient or rate limit is not reached.",
}

// The request was cancelled by the client.
var CANCELLED = responseData{
	Code:    499,
	Status:  "CANCELLED",
	Message: "Request cancelled by client",
}

// Irrecoverable data loss or data corruption
// The client should report errors to the user.
var DATA_LOSS = responseData{
	Code:    500,
	Status:  "DATA_LOSS",
	Message: "Error in processing data",
}

// Unknown server error, usually due to a bug in the server.
var UNKNOWN = responseData{
	Code:    500,
	Status:  "UNKNOWN",
	Message: "Server Unknown Error",
}

// Internal Server Error
var INTERNAL = responseData{
	Code:    500,
	Status:  "INTERNAL",
	Message: "Internal Server Error",
}

// API methods are not implemented by the server.
var NOT_IMPLEMENTED = responseData{
	Code:    501,
	Status:  "NOT_IMPLEMENTED",
	Message: "API does not exist",
}

// Service unavailable. This is usually due to server downtime.
var UNAVAILABLE = responseData{
	Code:    503,
	Status:  "UNAVAILABLE",
	Message: "service is not available",
}

// The request is past the deadline.
// This occurs only if the caller sets a deadline that is shorter than the default deadline for the method (the server was unable to process the request by the deadline) and the request did not complete within the deadline.
var DEALINE_EXCEED = responseData{
	Code:    504,
	Status:  "DEALINE_EXCEED",
	Message: "request timeout",
}

// Common status codes customized according to business
