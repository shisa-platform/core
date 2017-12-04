package service

type Endpoint struct {
	Method   string
	Route    string
	Policy   Policy
	Pipeline Pipeline
	// xxx - request (query|body) parameters
	// xxx - pipeline instead of handler
}
