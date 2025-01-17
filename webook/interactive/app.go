package main

import (
	"webook/internal/events"

	"github.com/to404hanga/pkg404/grpcx"
)

type App struct {
	consumers []events.Consumer
	server    *grpcx.Server
}
