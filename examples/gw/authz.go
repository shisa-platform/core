package main

import (
	"github.com/percolate/shisa/context"
)

type SimpleAuthorization struct{
	AllowedIDs []string
}

func (a SimpleAuthorization) Authorize(ctx context.Context, r *service.Request) (bool, merry.Error) {
	for _, id := range a.AllowedIDs {
		if ctx.Actor().ID() == id {
			return true, nil
		}
	}

	return false, nil
}

