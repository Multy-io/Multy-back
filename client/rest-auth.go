/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"net/http"
	"time"

	"github.com/Multy-io/Multy-back/store"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/mgo.v2/bson"
)

// The jwt middleware.

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
			restClient.middlewareJWT.unauthorized(c, http.StatusBadRequest, "Missing UserID, DeviceID, PushToken, AppVersion or DeviceType")
			return
		}

		if restClient.middlewareJWT.Authenticator == nil {
			restClient.middlewareJWT.unauthorized(c, http.StatusInternalServerError, "Missing define authenticator func")
			return
		}

		user, ok := restClient.middlewareJWT.Authenticator(loginVals.UserID, loginVals.DeviceID, loginVals.PushToken, loginVals.DeviceType, c) // user can be empty

		userID := loginVals.UserID

		// Create the token
		token := jwt.New(jwt.GetSigningMethod(restClient.middlewareJWT.SigningAlgorithm))
		claims := token.Claims.(jwt.MapClaims)
		if restClient.middlewareJWT.PayloadFunc != nil {
			for key, value := range restClient.middlewareJWT.PayloadFunc(loginVals.UserID) {
				claims[key] = value
			}
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

		if !ok {
			// new User with new Device
			device := createDevice(loginVals.DeviceID, c.ClientIP(), tokenString, loginVals.PushToken, loginVals.AppVersion, loginVals.DeviceType)

			var wallet []store.Wallet
			var devices []store.Device
			devices = append(devices, device)

			newUser := createUser(loginVals.UserID, devices, wallet)
			err = restClient.userStore.Insert(newUser)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"token":  "",
					"expire": "",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"token":  tokenString,
					"expire": expire.Format(time.RFC3339),
				})
			}
			return
		}

		// old user - check new or old device

		// If user auths with DeviceID and UserID that
		// already exists in DB we refresh JWT token.
		for _, concreteDevice := range user.Devices {
			// case of expired token on device or relogin from same device with
			// userID and deviceID existed in DB
			if concreteDevice.DeviceID == loginVals.DeviceID {
				restClient.log.Infof("update token for device %s", loginVals.DeviceID)
				sel := bson.M{"userID": user.UserID, "devices.JWT": concreteDevice.JWT}
				update := bson.M{"$set": bson.M{"devices.$.JWT": tokenString}}
				err = restClient.userStore.Update(sel, update)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"token":  "",
						"expire": "",
					})
				} else {
					c.JSON(http.StatusOK, gin.H{
						"token":  tokenString,
						"expire": expire.Format(time.RFC3339),
					})
				}
				return
			}
		}

		// no such device - creatig this one
		restClient.log.Infof("creating new device %s", loginVals.DeviceID)
		// case of adding new device to user account
		// e.g. user want to use app on another device
		device := createDevice(loginVals.DeviceID, c.ClientIP(), tokenString, loginVals.PushToken, loginVals.AppVersion, loginVals.DeviceType)
		user.Devices = append(user.Devices, device)

		sel := bson.M{"userID": userID}
		err = restClient.userStore.UpdateUser(sel, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"token":  "",
				"expire": "",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"token":  tokenString,
				"expire": expire.Format(time.RFC3339),
			})
		}
		return
	}
}
