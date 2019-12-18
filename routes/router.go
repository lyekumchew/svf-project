package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"svf-project/controllers"
)

func Init() *gin.Engine {
	router := gin.Default()

	// Default() allows all origins & add OPTIONS allows
	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	v1 := router.Group("/api/v1")
	{
		video := new(controllers.Video)
		v1.GET("/videos/:video_id", video.Index)
		v1.POST("/videos/:video_id", video.Show)
		v1.POST("/upload", video.Store)
		v1.POST("/delete/:delete_id", video.Destroy)
	}

	return router
}
