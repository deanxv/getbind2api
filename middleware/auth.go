package middleware

import (
	"getbind2api/common"
	"getbind2api/common/config"
	logger "getbind2api/common/loggger"
	"getbind2api/model"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"net/http"
	"strings"
)

func isValidSecret(secret string) bool {
	if config.ApiSecret == "" {
		return true
	} else {
		return lo.Contains(config.ApiSecrets, secret)
	}
}

func isValidBackendSecret(secret string) bool {
	return config.BackendSecret != "" && !(config.BackendSecret == secret)
}

func authHelperForOpenai(c *gin.Context) {
	secret := c.Request.Header.Get("Authorization")
	secret = strings.Replace(secret, "Bearer ", "", 1)

	b := isValidSecret(secret)

	if !b {
		c.JSON(http.StatusUnauthorized, model.OpenAIErrorResponse{
			OpenAIError: model.OpenAIError{
				Message: "API-KEY校验失败",
				Type:    "invalid_request_error",
				Code:    "invalid_authorization",
			},
		})
		c.Abort()
		return
	}

	//if config.ApiSecret == "" {
	//	c.Request.Header.Set("Authorization", "")
	//}

	c.Next()
	return
}

func authHelperForBackend(c *gin.Context) {
	secret := c.Request.Header.Get("Authorization")
	secret = strings.Replace(secret, "Bearer ", "", 1)
	if isValidBackendSecret(secret) {
		logger.Debugf(c.Request.Context(), "BackendSecret is not empty, but not equal to %s", secret)
		common.SendResponse(c, http.StatusUnauthorized, 1, "unauthorized", "")
		c.Abort()
		return
	}

	if config.BackendSecret == "" {
		c.Request.Header.Set("Authorization", "")
	}

	c.Next()
	return
}

func OpenAIAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelperForOpenai(c)
	}
}

func BackendAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelperForBackend(c)
	}
}
