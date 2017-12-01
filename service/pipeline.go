package service

import (
	"github.com/percolate/shisa/context"
)

type Handler func(context.Context, *Request) Response

type Pipeline []Handler
