package handles

import (
	"encoding/base64"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

type V2RayServer struct {
	Id       string `json:"id"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	Method   string `json:"method"`
	Password string `json:"password"`
}

type V2rayConfig struct {
	SubscribeUrls  []string `json:"subscribe_urls"`
	LastUpdateTime uint64   `json:"last_update_time"`
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

func UpdateSubscription(c *gin.Context) {
	conf_dir := conf.Conf.V2rayConfigDir
	v2ray_config := filepath.Join(conf_dir, "config.json")
	config_bytes, err := ioutil.ReadFile(v2ray_config)
	if err != nil {
		log.Print("Open v2rayConfig error!")
		common.ErrorResp(c, err, 500)
		return
	}
	var config V2rayConfig
	err = json.Unmarshal(config_bytes, &config)
	if err != nil {
		log.Print("Parse v2ray config error")
		common.ErrorResp(c, err, 500)
		return
	}

	config.LastUpdateTime = uint64(time.Now().Unix())
	servers := []V2RayServer{}
	for _, url := range config.SubscribeUrls {
		servers = append(servers, getServersFromUrl(url)...)
	}

	v2ray_server_config := filepath.Join(conf_dir, "servers.json")
	servers_json, err := json.Marshal(servers)
	if err != nil {
		log.Printf("encode servers error %v", err)
		common.ErrorResp(c, err, 500)
		return
	}
	os.Truncate(v2ray_server_config, 0)
	ioutil.WriteFile(v2ray_server_config, servers_json, fs.ModeAppend)
	common.SuccessResp(c, "update success")
}

func getServersFromUrl(urlString string) []V2RayServer {
	servers := []V2RayServer{}
	response, err := http.Get(urlString)
	if err != nil {
		log.Printf("Get for url=%v error %v", urlString, err)
		return servers
	}

	encodedServerInfo, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Read response from Get request for %v error %v", urlString, err)
		return servers
	}

	decodedServerInfo, err := base64.StdEncoding.DecodeString(string(encodedServerInfo))
	if err != nil {
		log.Printf("Decode Server info from remote error %v", err)
		return servers
	}

	serversStr := strings.Split(string(decodedServerInfo), "\n")
	for _, serverStr := range serversStr {
		parsedUrl, err := url.Parse(serverStr)
		if err != nil {
			log.Printf("Parse Server url=%v error %v", serverStr, err)
			continue
		}

		hostAndPort := strings.Split(parsedUrl.Host, ":")
		methodAndPasswordBytes, err := base64.StdEncoding.DecodeString(parsedUrl.User.Username())
		if err != nil {
			log.Printf("decode method and password error %v", err)
			continue
		}
		methodAndPassword := strings.Split(string(methodAndPasswordBytes), ":")

		server := V2RayServer{
			Address:  hostAndPort[0],
			Port:     hostAndPort[1],
			Method:   methodAndPassword[0],
			Password: methodAndPassword[1],
		}
		servers = append(servers, server)
	}

	return servers
}

func UpdateSubscribeUrls(c *gin.Context) {

}
