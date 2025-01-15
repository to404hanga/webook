package main

import (
	"webook/internal/events"

	"github.com/gin-gonic/gin"
	cron "github.com/robfig/cron/v3"
)

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
	cron      *cron.Cron
}
