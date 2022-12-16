package handles

import "github.com/gin-gonic/gin"

type V2RayServer struct {
	Id string `json:"id"`
	Address string `json:"address"`
	Port uint16 `json:"port"`
	Method string `json:"method"`
	Password string `json:"password"` 
}

func ListV2rayServers(c *gin.Context) {

}

func ApplyV2RayServer(c *gin.Context) {

}

func RestartV2RayServer(c *gin.Context) {

}

func StopV2rayServer(c *gin.Context) {

}

func SubscribeToUrl(c *gin.Context) {

}

func GetV2RayServerNetworkFlow(c *gin.Context) {

}

func ApplyV2rayRoute(c *gin.Context) {
	
}
