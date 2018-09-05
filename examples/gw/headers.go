package main

import (
	"time"

	"github.com/shisa-platform/core/httpx"
)

func addCommonHeaders(response httpx.Response) {
	now := time.Now().UTC().Format(time.RFC1123)

	response.Headers().Set("Cache-Control", "private, max-age=0")
	response.Headers().Set("Date", now)
	response.Headers().Set("Expires", now)
	response.Headers().Set("Last-Modified", now)
	response.Headers().Set("X-Content-Type-Options", "nosniff")
	response.Headers().Set("X-Frame-Options", "DENY")
	response.Headers().Set("X-Xss-Protection", "1; mode=block")
}
