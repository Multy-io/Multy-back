package client

import (
	"github.com/gin-gonic/gin"
)

func (restClient *RestClient) reportError(c *gin.Context, httpErrorCode int, httpMessage string, err error) {
	restClient.log.Errorf("Error: %v", err)
	c.JSON(httpErrorCode, gin.H{"code": httpErrorCode, "message": httpMessage})
}
