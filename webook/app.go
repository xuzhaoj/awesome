package main

import (
	"awesomeProject/webook/internal/events"
	"github.com/gin-gonic/gin"
)

type App struct {
	Server   *gin.Engine
	Consumer []events.Consumer
}
