package httpapi

import "github.com/HenryXiaoYang/powerschool-mirror/server/internal/student"

// Request and response bodies that are not domain models. Generated into web/src/api/types.gen.ts.

type LoginRequest struct {
	Username string `json:"username" binding:"required,max=256"`
	Password string `json:"password" binding:"required,max=256"`
}

type LoginResponse struct {
	Token string `json:"token"`
	// ReportCards is false when the portal sign-in failed; the rest of the app still works.
	ReportCards bool `json:"reportCards"`
}

type ReportCardList struct {
	ReportCards []student.ReportCard `json:"reportCards"`
}

// ErrorCode is the machine-readable error kind. session_expired means: sign in again and retry.
type ErrorCode string

const (
	CodeInvalidCredentials ErrorCode = "invalid_credentials"
	CodeSessionExpired     ErrorCode = "session_expired"
	CodeBadRequest         ErrorCode = "bad_request"
	CodeNotFound           ErrorCode = "not_found"
	CodeUpstream           ErrorCode = "upstream_error"
	CodeTimeout            ErrorCode = "timeout"
	CodeInternal           ErrorCode = "internal"
)

type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}
