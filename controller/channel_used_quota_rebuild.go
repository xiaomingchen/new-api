package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func RebuildChannelUsedQuota(c *gin.Context) {
	result, err := model.RebuildChannelUsedQuota()
	if err != nil {
		common.SysError("failed to rebuild channel used quota: " + err.Error())
		common.ApiError(c, err)
		return
	}
	common.SysLog("channel used quota rebuilt")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "channel used quota rebuilt",
		"data":    result,
	})
}
