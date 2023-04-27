package jsonerr

import (
	"encoding/json"
	"net/http"
)

// Error writes structured error information to w using JSON encoding.
// The given status code is used if it is non-zero, otherwise it
// is set to 500.
//
// If err is nil it writes sets the status to 200 OK and writes:
//
//	{"code": "ok", "message": ""}
func Error(w http.ResponseWriter, err error, code int) {
	if code == 0 {
		code = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if err == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
  "code": "ok",
  "message": ""
}
`))
		return
	}

	type Err struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	data, _ := json.MarshalIndent(&Err{
		Code:    http.StatusText(code),
		Message: err.Error(),
	}, "", "  ")
	w.WriteHeader(code)
	_, _ = w.Write(data)
}
