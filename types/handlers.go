package types

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var session map[int64]Members
var sessionLock sync.Mutex

type Members struct {
	UserID   int64 `gorm:"primaryKey;autoIncrement"`
	Nickname string
	Username string `gorm:"unique;index"`
	UserType UserType
	Password string
	DeleteAt gorm.DeletedAt
}

func DbInit() {
	// 连接至数据库
	var err error
	dsn := "root:bytedancecamp@tcp(127.0.0.1:3306)/ByteCamp?charset=utf8mb4&parseTime=True&loc=Local"
	//dsn := "root:SIhan1998!@tcp(127.0.0.1:3306)/ByteCamp?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("db open failed\n")
	}

	// init session

	session = make(map[int64]Members)

	// 检查默认管理员
	defaultAdmin := Members{
		UserID:   1,
		Nickname: "",
		Username: "JudgeAdmin",
		UserType: Admin,
		Password: "JudgePassword2022",
	}

	db.AutoMigrate(Members{})

	var admin *Members
	if err2 := db.Unscoped().Where("user_id = 1").First(&admin); err2.Error != nil {

		// 如果默认管理员不存在，添加至用户表
		result := db.Create(&defaultAdmin)

		if result.Error != nil {
			fmt.Println("create failed")
		}
	}

}

// CheckNickName check nickname validity
func CheckNickName(nickname string) bool {
	if len(nickname) < 4 || len(nickname) > 20 {
		return false
	}
	return true
}

func CheckUserName(userName string) bool {
	if len(userName) < 8 || len(userName) > 20 {
		return false
	}

	for _, r := range userName {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func CheckPassWord(passWord string) bool {
	hasUpper := false
	hasLower := false
	hasNumber := false

	if len(passWord) < 8 || len(passWord) > 20 {
		return false
	}

	for _, r := range passWord {
		if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsLower(r) {
			hasLower = true
		} else if unicode.IsNumber(r) {
			hasNumber = true
		} else {
			return false
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return false
	}

	return true
}

func CheckUserType(userType UserType) bool {
	if userType == Admin || userType == Student || userType == Teacher {
		return true
	}
	return false
}

func CreateMemberHandler(context *gin.Context) {
	var createReq CreateMemberRequest
	var createResp CreateMemberResponse
	if context.BindJSON(&createReq) == nil {

		cookie, err := context.Cookie("camp-session")
		userId, _ := strconv.ParseInt(cookie, 10, 64)

		sessionLock.Lock()
		usession, ok := session[userId]
		sessionLock.Unlock()

		if err != nil || !ok {
			// 用户未登陆
			createResp.Code = LoginRequired
			context.JSON(http.StatusOK, createResp)
			return
		}

		if usession.UserType != Admin {
			// 非管理员 无操作权限
			createResp.Code = PermDenied
			context.JSON(http.StatusOK, createResp)
			return
		}

		//检查参数合法性
		if !CheckNickName(createReq.Nickname) || !CheckUserName(createReq.Username) || !CheckPassWord(createReq.Password) || !CheckUserType(createReq.UserType) {
			createResp.Code = ParamInvalid
			context.JSON(http.StatusOK, createResp)
			return
		}

		// 添加成员
		memberToAdd := Members{
			Nickname: createReq.Nickname,
			Username: createReq.Username,
			UserType: createReq.UserType,
			Password: createReq.Password,
		}
		result := db.Create(&memberToAdd)
		if result.Error != nil {
			// user already exist
			// fmt.Println("create failed")
			createResp.Code = UserHasExisted
			context.JSON(http.StatusOK, createResp)
			return
		}

		createResp.Data.UserID = strconv.FormatInt(memberToAdd.UserID, 10)
		createResp.Code = OK
		context.JSON(http.StatusOK, createResp)
	}
}

func GetOneMemberHandler(context *gin.Context) {
	var getMemberReq GetMemberRequest
	var getMemberResp GetMemberResponse

	if context.BindJSON(&getMemberReq) == nil {

		//todo: 可以尝试先在session里找，找到后直接返回，但是会增加情况的复杂度

		var member *Members
		usrID, _ := strconv.ParseInt(getMemberReq.UserID, 10, 64)

		if err2 := db.Unscoped().Where("user_id = ?", usrID).First(&member); err2.Error != nil {
			// 用户不存在
			getMemberResp.Code = UserNotExisted
			context.JSON(http.StatusOK, getMemberResp)
			return
		}

		if member.DeleteAt.Valid {
			// 用户已删除
			getMemberResp.Code = UserHasDeleted
			context.JSON(http.StatusOK, getMemberResp)
			return
		}

		getMemberResp.Data.Nickname = member.Nickname
		getMemberResp.Data.UserID = getMemberReq.UserID
		getMemberResp.Data.Username = member.Username
		getMemberResp.Data.UserType = member.UserType

		context.JSON(http.StatusOK, getMemberResp)
	}
}

func GetMemberListHandler(context *gin.Context) {
	var getListReq GetMemberListRequest
	var getListResp GetMemberListResponse

	if context.BindJSON(&getListReq) == nil {
		var members []Members
		db.Where("user_id >= ?", 0).Limit(getListReq.Limit).Offset(getListReq.Offset).Find(&members)

		for _, mem := range members {
			res := TMember{
				UserID:   strconv.FormatInt(mem.UserID, 10),
				Nickname: mem.Nickname,
				Username: mem.Username,
				UserType: mem.UserType,
			}

			getListResp.Data.MemberList = append(getListResp.Data.MemberList, res)
		}

		getListResp.Code = OK
		context.JSON(http.StatusOK, getListResp)
	}
}

func UpdateMemberHandler(context *gin.Context) {
	// todo: 记得updatemember后更新session
	var updateReq UpdateMemberRequest
	var updateResp UpdateMemberResponse

	if context.BindJSON(&updateReq) == nil {
		var member *Members
		usrID, _ := strconv.ParseInt(updateReq.UserID, 10, 64)

		if !CheckNickName(updateReq.Nickname) {
			updateResp.Code = ParamInvalid
			context.JSON(http.StatusOK, updateResp)
			return
		}

		if err2 := db.Unscoped().Where("user_id = ?", usrID).First(&member); err2.Error != nil {
			// 用户不存在
			updateResp.Code = UserNotExisted
			context.JSON(http.StatusOK, updateResp)
			return
		}

		if member.DeleteAt.Valid {
			// 用户已删除
			updateResp.Code = UserHasDeleted
			context.JSON(http.StatusOK, updateResp)
			return
		}

		db.Model(member).Update("nickname", updateReq.Nickname)
		updateResp.Code = OK
		context.JSON(http.StatusOK, updateResp)
		// todo: update session

		sessionLock.Lock()
		tmp := session[usrID]
		tmp.Nickname = updateReq.Nickname
		session[usrID] = tmp
		sessionLock.Unlock()
		return
	}

}

func DeleteMemberHandler(context *gin.Context) {
	//todo: 记得deletemember后更新session
	var deleteReq DeleteMemberRequest
	var deleteResp DeleteMemberResponse

	if context.BindJSON(&deleteReq) == nil {
		var member *Members
		usrID, _ := strconv.ParseInt(deleteReq.UserID, 10, 64)

		if err2 := db.Unscoped().Where("user_id = ?", usrID).First(&member); err2.Error != nil {
			// 用户不存在
			deleteResp.Code = UserNotExisted
			context.JSON(http.StatusOK, deleteResp)
			return
		}

		if member.DeleteAt.Valid {
			// 用户已删除
			deleteResp.Code = UserHasDeleted
			context.JSON(http.StatusOK, deleteResp)
			return
		}

		db.Delete(&member)
		deleteResp.Code = OK
		context.JSON(http.StatusOK, deleteResp)
		return
	}
}

func UpdateSession(member *Members) {
	sessionLock.Lock()
	defer sessionLock.Unlock()
	session[member.UserID] = *member
}

// 处理用户登入
func LogInHandler(context *gin.Context) {
	var loginReq LoginRequest
	var loginResp LoginResponse

	if err := context.BindJSON(&loginReq); err == nil {

		var member *Members

		if err2 := db.Unscoped().Where("username = ?", loginReq.Username).First(&member); err2.Error != nil {
			// 用户不存在
			loginResp.Code = WrongPassword
			context.JSON(http.StatusOK, loginResp)
			return
		}

		if member.DeleteAt.Valid {
			// 用户已删除
			loginResp.Code = WrongPassword
			context.JSON(http.StatusOK, loginResp)
			return
		}

		// todo: 检查若用户已登陆？

		if loginReq.Password == member.Password {
			// 更新session
			UpdateSession(member)

			// 返回登陆成功
			loginResp.Data.UserID = strconv.FormatInt(member.UserID, 10)
			loginResp.Code = 0
			context.SetCookie("camp-session", loginResp.Data.UserID, 5*365*24*60*60, "/",
				"180.184.67.164", false, true)
			context.JSON(http.StatusOK, loginResp)
			return
		}

		//密码错误
		loginResp.Code = WrongPassword
		context.JSON(http.StatusOK, loginResp)
	}
}

// 处理用户登出
func LogOutHandler(context *gin.Context) {
	var logOutResp LogoutResponse

	cookie, err := context.Cookie("camp-session")
	userId, _ := strconv.ParseInt(cookie, 10, 64)

	sessionLock.Lock()
	_, ok := session[userId]
	sessionLock.Unlock()

	if err != nil || !ok {
		// 用户未登陆
		logOutResp.Code = LoginRequired
		context.JSON(http.StatusOK, logOutResp)
		return
	}

	// 删除session
	// sessionLock.Lock()
	// delete(session, userId)
	// sessionLock.Unlock()

	// 删除cookie
	context.SetCookie("camp-session", "CookieNotSet", -1, "/", "180.184.67.164", false, true)

	logOutResp.Code = OK
	context.JSON(http.StatusOK, logOutResp)
}

func WhoAmIHandler(context *gin.Context) {
	member, ok := GetSession(context)

	var resp WhoAmIResponse

	if !ok {
		// 用户未登陆
		resp.Code = LoginRequired
		context.JSON(http.StatusOK, resp)
		return
	}

	resp.Code = OK
	resp.Data.Nickname = member.Nickname
	resp.Data.UserID = strconv.FormatInt(member.UserID, 10)
	resp.Data.Username = member.Username
	resp.Data.UserType = member.UserType
	context.JSON(http.StatusOK, resp)
}

func GetSession(context *gin.Context) (Members, bool) {
	cookie, err := context.Cookie("camp-session")
	userId, _ := strconv.ParseInt(cookie, 10, 64)
	sessionLock.Lock()
	member, ok := session[userId]
	sessionLock.Unlock()

	if err != nil || !ok {
		// fmt.Printf("getting session failed uid : %d", userId)
		return Members{}, false
	}

	return member, true
}
