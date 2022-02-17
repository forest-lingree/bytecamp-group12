package types

import "github.com/gin-gonic/gin"

// import "byteCamp/handlers"

func RegisterRouter(r *gin.Engine) {
	g := r.Group("/api/v1")

	// 成员管理
	g.POST("/member/create", CreateMemberHandler)
	g.GET("/member", GetOneMemberHandler)
	g.GET("/member/list", GetMemberListHandler)
	g.POST("/member/update", UpdateMemberHandler)
	g.POST("/member/delete", DeleteMemberHandler)

	// 登录

	g.POST("/auth/login", LogInHandler)
	g.POST("/auth/logout", LogOutHandler)
	g.GET("/auth/whoami", WhoAmIHandler)

	// 排课
	g.POST("/course/create")
	g.GET("/course/get")

	g.POST("/teacher/bind_course")
	g.POST("/teacher/unbind_course")
	g.GET("/teacher/get_course")
	g.POST("/course/schedule")

	// 抢课
	g.POST("/student/book_course")
	g.GET("/student/course")

}
