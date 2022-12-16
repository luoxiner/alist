package handles

import "github.com/gin-gonic/gin"

type ResourceUsage struct {
	 DiskUsed uint64 `json:"disk-Used"`
	 DiskTotal uint64 `json:"disk_total"`
}

func GetSystemResourceUsage(c *gin.Context) {
	
}