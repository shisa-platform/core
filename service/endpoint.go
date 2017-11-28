package service

type Endpoint struct {
	Method  string
	Route   string
	Policy  Policy
	Handler Handler
	// xxx - request (query|body) parameters
	// xxx - pipeline instead of handler
}
