package routes

import (
	"github.com/gin-gonic/gin"

	"siakad-mini/controllers"
	"siakad-mini/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/v1")

	api.POST("/auth/login", controllers.Login)

	auth := api.Group("")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("/auth/me", controllers.Me)

		adminOnly := auth.Group("/students")
		adminOnly.Use(middleware.RoleOnly("admin"))
		{
			adminOnly.GET("", controllers.GetStudents)
			adminOnly.POST("", controllers.CreateStudent)
			adminOnly.PUT("/:id", controllers.UpdateStudent)
			adminOnly.DELETE("/:id", controllers.DeleteStudent)
		}

		auth.GET("/students/:id", controllers.GetStudentByID)
		auth.GET("/courses", controllers.GetCourses)

		enroll := auth.Group("/enrollments")
		enroll.Use(middleware.RoleOnly("mahasiswa"))
		{
			enroll.POST("", controllers.CreateEnrollment)
			enroll.DELETE("/:id", controllers.DeleteEnrollment)
		}
	}

	return r
}
