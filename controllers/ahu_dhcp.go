package controllers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type Ahunet struct {
	Username string
	Password string
	base     string
	client   *http.Client
}

// return the ahunet struct
func NewAhuNet(username, password string) Ahunet {
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
	return Ahunet{
		Username: username,
		Password: password,
		base:     "http://172.16.253.3",
		client:   client,
	}
}

// get the ipv4 address info
func (ahu *Ahunet) GetIpv4Info() string {
	resp, err := ahu.client.Get(ahu.base + "/drcom/chkstatus?callback=dr1002&v=123")
	if err != nil {
		log.Println("get info error:", err)
		return ""
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("read stream body err:", err)
		return ""
	}
	s := dealJsonP(*(*string)(unsafe.Pointer(&bs)))
	ipv4 := gjson.Get(s, "v46ip").String()
	return ipv4
}

// authentication
func (ahu *Ahunet) Auth(ipv4 string) error {
	q := url.Values{}
	q.Add("c", "Portal")
	q.Add("a", "login")
	q.Add("callback", "dr1003")
	q.Add("login_method", "1")
	q.Add("user_account", ahu.Username)
	q.Add("user_password", ahu.Password)
	q.Add("wlan_user_ip", ipv4)
	q.Add("wlan_user_ipv6", "")
	q.Add("wlan_user_mac", "000000000000")
	q.Add("wlan_ac_ip", "")

	resp, err := ahu.client.Get(ahu.base + ":801/eportal/?" + q.Encode())
	if err != nil {
		log.Println("auth error:", err)
		return err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("read stream body err:", err)
		return err
	}
	s := dealJsonP(*(*string)(unsafe.Pointer(&bs)))
	result := gjson.Parse(s)
	code := result.Get("ret_code").Int()

	if code == 1 {
		log.Println("认证成功! IP:", ipv4)
		return nil
	}
	if code == 2 {
		log.Println("已经在线! IP:", ipv4)
		return errors.New("already online")
	}

	return errors.New(result.Get("msg").String())
}

// trim left bracket and right bracket
func dealJsonP(origin string) string {
	left := strings.Index(origin, "(")
	right := strings.LastIndex(origin, ")")
	return origin[left+1 : right]
}

func AhuDchpAddressAuth(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	if username == "" || password == "" {
		c.String(http.StatusOK, "username or password is empty")
		return
	}
	ahu := NewAhuNet(username, password)
	err := ahu.Auth(ahu.GetIpv4Info())
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	c.String(http.StatusOK, "auth success")
}
