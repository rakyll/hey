package requester

import "net/http"

// RequestFactory should be implemented to provide specific logic for creation of the requests
type RequestFactory interface {
	Create(i int) (*http.Request, error)
}
