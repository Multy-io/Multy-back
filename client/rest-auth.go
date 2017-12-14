package client

import (
	"net/http"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/mgo.v2/bson"
)

// The jwt middleware.

/*
var authenticator = func (restClient *RestClient) getAdressBalance() gin.HandlerFunc {
	return func(c *gin.Context) {


func(userId string, deviceId string, password string, c *gin.Context) (store.User, bool) {

	query := bson.M{"userID": userId}

	user := store.User{}

	err := usersData.Find(query).One(&user)

	if err != nil || len(user.UserID) == 0 {
		return store.User{}, false
	}

	return user, true

}
*/
// var authorizator = func(userId string, c *gin.Context) bool {
// 	if userId == "admin" {
// 		return true
// 	}
//
// 	return false
// }
/*
var unauthorized = func(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}
*/
// LoginHandler can be used by clients to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD", "deviceid": "DEVICEID"}.
// Reply will be of the form {"token": "TOKEN"}.
// func (mw *GinJWTMiddleware) LoginHandler(c *gin.Context) {
func (restClient *RestClient) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Initial middleware default setting.
		restClient.middlewareJWT.MiddlewareInit()

		var loginVals Login

		if c.ShouldBindWith(&loginVals, binding.JSON) != nil {
			restClient.middlewareJWT.unauthorized(c, http.StatusBadRequest, "Missing UserID, DeviceID, PushToken or DeviceType")
			return
		}

		if restClient.middlewareJWT.Authenticator == nil {
			restClient.middlewareJWT.unauthorized(c, http.StatusInternalServerError, "Missing define authenticator func")
			return
		}

		user, ok := restClient.middlewareJWT.Authenticator(loginVals.UserID, loginVals.DeviceID, loginVals.PushToken, loginVals.DeviceType, c) // user can be empty

		userID := user.UserID

		if len(userID) == 0 {
			ok = false
		}

		// Create the token
		token := jwt.New(jwt.GetSigningMethod(restClient.middlewareJWT.SigningAlgorithm))
		claims := token.Claims.(jwt.MapClaims)
		if restClient.middlewareJWT.PayloadFunc != nil {
			for key, value := range restClient.middlewareJWT.PayloadFunc(loginVals.UserID) {
				claims[key] = value
			}
		}

		if userID == "" {
			userID = loginVals.UserID
		}

		expire := restClient.middlewareJWT.TimeFunc().Add(restClient.middlewareJWT.Timeout)
		claims["id"] = userID
		claims["exp"] = expire.Unix()
		claims["orig_iat"] = restClient.middlewareJWT.TimeFunc().Unix()

		tokenString, err := token.SignedString(restClient.middlewareJWT.Key)
		if err != nil {
			restClient.middlewareJWT.unauthorized(c, http.StatusUnauthorized, "Create JWT Token faild")
			return
		}

		// If user auths with DeviceID and UserID that
		// already exists in DB we refresh JWT token.
	loop:
		for _, concreteDevice := range user.Devices {

			// case of expired token on device or relogin from same device with
			// userID and deviceID existed in DB
			if concreteDevice.DeviceID == loginVals.DeviceID {
				sel := bson.M{"userID": user.UserID, "devices.JWT": concreteDevice.JWT}
				update := bson.M{"$set": bson.M{"devices.$.JWT": tokenString}}
				err = restClient.userStore.Update(sel, update)
				responseErr(c, err, http.StatusInternalServerError) // 500
				break loop
			} else {
				// case of adding new device to user account
				// e.g. user want to use app on another device
				device := createDevice(loginVals.DeviceID, c.ClientIP(), tokenString, loginVals.PushToken, loginVals.DeviceType)
				user.Devices = append(user.Devices, device)

				sel := bson.M{"userID": userID}
				err = restClient.userStore.UpdateUser(sel, &user)

				responseErr(c, err, http.StatusInternalServerError) // 500
				break loop
			}
		}

		if !ok {
			device := createDevice(loginVals.DeviceID, c.ClientIP(), tokenString, loginVals.PushToken, loginVals.DeviceType)

			var wallet []store.Wallet
			var devices []store.Device
			devices = append(devices, device)

			newUser := createUser(loginVals.UserID, devices, wallet)

			user = newUser

			err = restClient.userStore.Insert(user)
			responseErr(c, err, http.StatusInternalServerError) // 500

		}

		c.JSON(http.StatusOK, gin.H{
			"token":  tokenString,
			"expire": expire.Format(time.RFC3339),
		})
	}
}
