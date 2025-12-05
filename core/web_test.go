package core

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"testing"
	"time"
)

func TestCors(t *testing.T) {

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  false,
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST"},
		AllowCredentials: true,
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool {
			log.Println("=========origin", origin) // 这个会打印
			return true
		},
		// AllowOriginFunc: func(origin string) bool {  // 注释掉，避免冲突
		//     log.Println("origin", origin)
		//     return true
		// },
		MaxAge: 12 * time.Hour,
	}))

	router.POST("/api/sendValidateCode", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World!"})
	})
	router.GET("/api/sendValidateCode", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World!"})
	})

	if err := router.Run(":8081"); err != nil {
		panic(err)
	}

}
