package route

import (
	"ginmongo/controller"

	"github.com/gin-gonic/gin"
)

func Unprotected(router *gin.Engine) {
	router.POST("/registration", controller.RegisterUser)
	router.POST("/login", controller.Login)
	router.POST("/logout", controller.Logout)
	router.POST("/forgetpassword", controller.ForgetPassword)
	router.POST("/users/resetpassword/reset/:resetcode", controller.ResetPassword)
	router.POST("/users/:user_id", controller.UpdatePassword)

}
