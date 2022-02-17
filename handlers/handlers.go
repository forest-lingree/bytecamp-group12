package types

import (
	"byteCamp/types"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

var db *gorm.DB
var usrPassWord map[string]string // need to be changed to redis?

func DbInit() {
	var err error
	db, err = gorm.Open(mysql.Open("byteCamp"), &gorm.Config{})

	if err != nil {
		fmt.Println("Db init failed")
		return
	}

	//todo: change this to redis
	usrPassWord["JudgeAdmin"] = "JudgePassword2022"
}

// CheckNickName check nickname validity
func CheckNickName(nickname string) bool {
	if len(nickname) < 4 || len(nickname) > 20 {
		return false
	}
	return true
}

func CreateMemberHandler(context *gin.Context) {
	var createReq types.CreateMemberRequest
	var createResp types.CreateMemberResponse

	if context.BindJSON(&createReq) == nil {

		// todo: check usr cookie : login, isAdmin
		// warning: need to add lock to this usr during this check

		if !CheckNickName(createReq.Nickname) {
			createResp.Code = types.ParamInvalid
			context.JSON(http.StatusOK, createResp)
			return
		}

		// ...
	}
}

func LogInHandler(context *gin.Context) {
	var loginReq types.LoginRequest
	var loginResp types.LoginResponse

	if context.BindJSON(&loginReq) == nil {
		if usrPassWord[loginReq.Username] == loginReq.Password {
			loginResp.Data.UserID = "1"
			loginResp.Code = 0
			context.JSON(http.StatusOK, loginResp)
			return
		}

		loginResp.Data.UserID = "0"
		loginResp.Code = types.WrongPassword
		context.JSON(http.StatusOK, loginResp)
	}
}
