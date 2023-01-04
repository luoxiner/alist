package handles

import (
	"encoding/base64"
	"encoding/json"
	"errors"
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
	"github.com/alist-org/alist/v3/pkg/utils"
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

type V2rayLog struct {
	Access   string `json:"access"`
	Error    string `json:"error"`
	LogLevel string `json:"loglevel"`
}

type V2rayInbound struct {
	Port     int         `json:"port"`
	Listen   string      `json:"listen"`
	Protocol string      `json:"protocol"`
	Settings interface{} `json:"settings"`
	Tag      string      `json:"tag"`
}

type V2rayOutbound struct {
	Protocol    string      `json:"protocol"`
	SendThrough string      `json:"sendThrough"`
	Settings    interface{} `json:"settings"`
	Tag         string      `json:"tag"`
}

type V2rayDns struct {
	Servers []string `json:"servers"`
}

type V2rayApi struct {
	Services []string `json:"services"`
	Tag      string   `json:"tag"`
}

type V2rayPolicy struct {
	System map[string]bool `json:"systems"`
}

type V2rayRouting struct {
	DomainStrategy string `json:"domainStrategy"`
}

type V2rayRoutingRule struct {
	InboundTag  []string `json:"inboundTag"`
	OutboundTag string   `json:"outboundTag"`
	Type        string   `json:"type"`
	Domain      []string `json:"domain"`
	Ip          []string `json:"ip"`
	Port        string   `json:"port"`
	Network     string   `json:"network"`
	Source      []string `json:"source"`
	User        []string `json:"user"`
	Protocol    []string `json:"protocol"`
	Attrs       string   `json:"attrs"`
	BalancerTag string   `json:"balancerTag"`
}

type V2rayRoutingBalancer struct {
	Tag      string   `json:"tag"`
	Selector []string `json:"selector"`
}

type V2rayServerConfig struct {
	Log       V2rayLog        `json:"log"`
	Api       V2rayApi        `json:"api"`
	Dns       V2rayDns        `json:"dns"`
	Routing   V2rayRouting    `json:"routing"`
	Policy    V2rayPolicy     `json:"policy"`
	Inbounds  []V2rayInbound  `json:"inbounds"`
	Outbounds []V2rayOutbound `json:"outbounds"`
}

func ListV2rayServers(c *gin.Context) {
	conf_dir := conf.Conf.V2rayConfigDir
	v2ray_servers_file := filepath.Join(conf_dir, "servers.json")
	_, err := os.Stat(v2ray_servers_file)
	if os.IsNotExist(err) {
		common.ErrorResp(c, err, 500)
		return
	}

	servers_bytes, err := ioutil.ReadFile(v2ray_servers_file)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	servers := []V2RayServer{}
	err = json.Unmarshal(servers_bytes, &servers)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, servers)
}

func ApplyV2RayServer(c *gin.Context) {
	var req V2RayServer
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	config, err := loadV2rayServerConfigFromFile()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	settings := []map[string]interface{}{
		{
			"address":  req.Address,
			"method":   req.Method,
			"ota":      false,
			"password": req.Password,
			"port":     req.Port,
		},
	}

	for index := range config.Outbounds {
		if config.Outbounds[index].Protocol == "shadowsocks" {
			config.Outbounds[index].Settings = settings
		}
	}

	saveV2rayServerConfigTo()
	err = restartV2rayServer()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, "ok")
}

func RestartV2RayServer(c *gin.Context) {
	err := restartV2rayServer()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	common.SuccessResp(c, "ok")
}

func restartV2rayServer() error {
	docker, err := utils.GetDefaultDocker()
	if err != nil {
		return err
	}

	containers, err := docker.ListContainers()
	if err != nil {
		return err
	}
	for _, container := range containers {
		if container.Image == "" {
			err := docker.RestartContainer(container.ID)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return errors.New("v2ray container not found")
}

func StopV2rayServer(c *gin.Context) {

}

func loadV2rayServerConfigFromFile(path string) (*V2rayServerConfig, error) {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(V2rayServerConfig)
	err = json.Unmarshal(configBytes, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func saveV2rayServerConfigTo(path string, config V2rayServerConfig) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.Truncate(path, 0)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, bytes, fs.ModeAppend)
	if err != nil {
		return err
	}

	return nil
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
			Id:       parsedUrl.Host,
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
