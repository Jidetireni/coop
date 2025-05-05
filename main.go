package main

import (
	config "cooperative-system/conf"
	"cooperative-system/routers"
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
