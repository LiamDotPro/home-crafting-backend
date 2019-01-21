package main

import "github.com/gin-gonic/gin"

func SetupUserRoutes(Router *gin.Engine) {
	// Users
	setupUsersRoutes(Router)

}

func SetupMasterRoutes(Router *gin.Engine) {

	// Master Users
	setupMasterUsersRoutes(Router)

}
