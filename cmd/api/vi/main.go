package main

import (
	"cooperative-system/internal/config"
	"cooperative-system/internal/routers"
	"os"

	"github.com/gin-gonic/gin"
)

func init() {
	config.LoadEnvVars()
	config.ConnectDb()
	config.SyncDB()

}

func main() {

	r := gin.Default()
	routers.SetUpRoute(r)
	r.Run(":" + os.Getenv("PORT"))
}
