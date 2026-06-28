package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/ccapi/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterMediaRoutes serves generated media saved through the shared local
// object storage configuration. Remote object stores should return their own
// signed/public URLs and do not use this route.
func RegisterMediaRoutes(v1 *gin.RouterGroup, settingService *service.SettingService) {
	if v1 == nil {
		return
	}
	v1.GET("/media/*key", func(c *gin.Context) {
		if settingService == nil {
			c.Status(http.StatusNotFound)
			return
		}
		key := strings.TrimLeft(c.Param("key"), "/")
		path, err := settingService.ResolveLocalMediaPath(c.Request.Context(), key)
		if err != nil {
			if errors.Is(err, service.ErrBackupS3NotConfigured) {
				c.Status(http.StatusNotFound)
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"type": "invalid_request_error", "message": "invalid media key"}})
			return
		}
		c.File(path)
	})
}
