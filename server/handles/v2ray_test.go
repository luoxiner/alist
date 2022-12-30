package handles

import (
	"net/url"
	"testing"
)

func Test_getServersFromUrl(t *testing.T) {
	parseUrl, _ := url.Parse("ss://YWVzLTI1Ni1nY206ZjUwMDgyNGYtYzA5ZC00ODM3LWFkMGItOWVmZGNhNDQxOWQ5@iepl.teacat2.xyz:49991#IEPL%20%E9%A6%99%E6%B8%AF%20%20%E2%82%82.%E2%82%85%E2%82%93")
	println(parseUrl.Host)
	println(parseUrl.Scheme)
	println(parseUrl.User.Username())
}
