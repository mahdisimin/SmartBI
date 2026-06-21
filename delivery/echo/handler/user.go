package echowebframework

import (
	"intelligentBI/repository/SQLServer"
	"intelligentBI/service"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

func UserRegisterHandler(c *echo.Context) error {
	var userRegReq service.UserRegisterRequest
	var userRegResp service.UserRegisterResponse
	var userServ = service.UserService{
		Repository: SQLServer.SqlServer{},
	}

	if err := c.Bind(&userRegReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userRegRespTemp, err := userServ.Register(userRegReq)

	if err != nil {
		log.Printf("error on HTTP request , register error : %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	} else {
		userRegResp = userRegRespTemp
	}
	return c.JSON(http.StatusOK, userRegResp)
}

func UserLoginHandler(c *echo.Context) error {
	var userLoginReq service.UserLoginRequest
	var userLoginResp service.UserLoginResponse
	var userServ = service.UserService{
		Repository: SQLServer.SqlServer{},
	}

	if err := c.Bind(&userLoginReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userLoginRespTemp, err := userServ.Login(userLoginReq)
	if err != nil {
		log.Printf("error on HTTP request , login error : %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())

	} else {
		userLoginResp = userLoginRespTemp
	}
	return c.JSON(http.StatusOK, userLoginResp)

}

func UserProfileHandler(c *echo.Context) error {

	var userProfileResp service.UserProfileResponse
	var userServ = service.UserService{
		Repository: SQLServer.SqlServer{},
	}

	userIdStr := c.Param("id")
	userId, err := strconv.Atoi(userIdStr)

	userProfileRequest := service.UserProfileRequest{
		UserID: int64(userId),
	}
	userProfileRespTemp, err := userServ.Profile(userProfileRequest)
	if err != nil {
		log.Printf("error on HTTP request , call service : %v", err)
		echo.NewHTTPError(http.StatusBadRequest, err.Error())
	} else {
		userProfileResp = userProfileRespTemp
	}

	return c.JSON(http.StatusOK, userProfileResp)
}
