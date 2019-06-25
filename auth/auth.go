package auth

import "net/http"

// FailedResponse represents the result of a failed authentication or authorization step
type FailedResponse struct {
	Message    string
	StatusCode int
}

// DoWrite writes the representation of the response to the http writer.
// The writer will close the response stream, so nothing more should be written to the ResponseWriter after this method is called
func (r *FailedResponse) DoWrite(w http.ResponseWriter) {
	w.WriteHeader(r.StatusCode)
	w.Write([]byte(r.Message))
	return
}
