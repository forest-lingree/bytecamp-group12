package main

import (
	// "byteCamp/handlers"
	"byteCamp/types"
	"github.com/gin-gonic/gin"
)

func main() {
	g := gin.Default()
	types.DbInit()
	types.RegisterRouter(g)

	g.Run(":80")
}
