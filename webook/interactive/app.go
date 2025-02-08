package main

import (
	"webook/interactive/events"

	"github.com/to404hanga/pkg404/ginx"
	"github.com/to404hanga/pkg404/grpcx"
)

type App struct {
	consumers   []events.Consumer
	server      *grpcx.Server
	adminServer *ginx.Server
}
