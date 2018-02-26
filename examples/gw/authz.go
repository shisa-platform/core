package main

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

type SimpleAuthorization struct {
	AllowedIDs []string
}

func (a SimpleAuthorization) Authorize(ctx context.Context, r *httpx.Request) (bool, merry.Error) {
	for _, id := range a.AllowedIDs {
		if ctx.Actor().ID() == id {
			return true, nil
		}
	}

	return false, nil
}
