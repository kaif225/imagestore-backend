package route

import (
	"ginmongo/controller"
	mw "ginmongo/middlewares"

	"github.com/gin-gonic/gin"
)

func Protected(router *gin.Engine) {

	protected := router.Group("/")

	protected.Use(mw.JWT())
	protected.POST("/upload/:category", controller.UploadImage)
	protected.GET("/images/:category", controller.GetImagesByCategory)
	protected.GET("/images", controller.GetAllImages)
	protected.GET("/images/search", controller.GetImagesByName)
	protected.GET("/category", controller.GetCategories)
	protected.POST("/category", controller.CreateCategory)
}
