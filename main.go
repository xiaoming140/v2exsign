package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	for i := 0; i < 3; i++ {
		cansign, err := check()
		if err != nil {
			log.Println(err)
			continue
		}
		if cansign {
			once, err := getonce()
			if err != nil {
				log.Println(err)
				continue
			}
			_, err = httpget(`https://www.v2ex.com` + once)
			if err != nil {
				log.Println(err)
				continue
			}
			cansign, err := check()
			if err != nil {
				log.Println(err)
				continue
			}
			if cansign {
				log.Println("签到失败，尝试重签")
				continue
			}
			i, err := getbalance()
			if err != nil {
				log.Println(err)
				continue
			}
			msg := "签到成功，本次签到获得 " + strconv.Itoa(i) + " 铜币。"
			log.Println(msg)
			if sckey != "" {
				for i := 0; i < 3; i++ {
					err := push(msg, sckey)
					if err != nil {
						log.Println(err)
						continue
					}
				}
			}
			return
		}
		log.Println("签过到了")
		return
	}
	panic("签到失败")
}

func getonce() (string, error) {
	b, err := httpget("https://www.v2ex.com/mission/daily")
	if err != nil {
		return "", fmt.Errorf("getonece: %w", err)
	}
	once := oncereg.Find(b)
	if once == nil {
		return "", &NotFind{msg: string(b)}
	}
	one := string(once)
	log.Println(one)
	return one, nil
}

func check() (bool, error) {
	b, err := httpget("https://www.v2ex.com/mission/daily")
	if err != nil {
		return false, fmt.Errorf("check: %w", err)
	}
	if bytes.Contains(b, []byte(`需要先登录`)) {
		panic("cookie 失效")
	}
	if bytes.Contains(b, []byte(`每日登录奖励已领取`)) {
		return false, nil
	}
	return true, nil
}

var (
	c      = http.Client{Timeout: 5 * time.Second}
	cookie string
	sckey  string
)

func init() {
	cookie = os.Getenv("v2exCookie")
	if cookie == "" {
		panic("你 cookie 呢？")
	}
	sckey = os.Getenv("sckey")
}

func httpget(url string) ([]byte, error) {
	reqs, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("httpget: %w", err)
	}
	reqs.Header.Set("Accept", "*/*")
	reqs.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 10; GM1900) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.86 Mobile Safari/537.36")
	reqs.Header.Set("Cookie", cookie)
	rep, err := c.Do(reqs)
	if rep != nil {
		defer rep.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("httpget: %w", err)
	}
	if rep.StatusCode != http.StatusOK {
		return nil, Errpget{msg: rep.Status, url: url}
	}
	b, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		return nil, fmt.Errorf("httpget: %w", err)
	}
	return b, nil
}

type Errpget struct {
	msg string
	url string
}

func (h Errpget) Error() string {
	return "not 200: " + h.msg + " " + h.url
}

var (
	balancereg = regexp.MustCompile(`的每日登录奖励 ([0-9]{1,4}) 铜币`)
	oncereg    = regexp.MustCompile(`/mission/daily/redeem\?once=\d+`)
)

func getbalance() (int, error) {
	b, err := httpget(`https://www.v2ex.com/balance`)
	if err != nil {
		return 0, fmt.Errorf("getbalance: %w", err)
	}
	temp := balancereg.FindSubmatch(b)
	if temp == nil {
		return 0, &NotFind{string(b)}
	}
	i, err := strconv.ParseInt(string(temp[1]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("getbalance: %w", err)
	}
	return int(i), nil
}

type NotFind struct {
	msg string
}

func (n *NotFind) Error() string {
	return "not find in: " + n.msg
}

func push(msg, key string) error {
	msg = `text=v2ex签到信息&desp=` + url.QueryEscape(msg)
	reqs, err := http.NewRequest("POST", "https://sc.ftqq.com/"+key+".send", strings.NewReader(msg))
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	reqs.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rep, err := c.Do(reqs)
	if rep != nil {
		defer rep.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	b, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	e := Returnmsg{}
	err = json.Unmarshal(b, &e)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	if e.Errno != 0 {
		return e
	}
	return nil
}

//{"errno":0,"errmsg":"success","dataset":"done"}

type Returnmsg struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
}

func (r Returnmsg) Error() string {
	return "code: " + strconv.Itoa(r.Errno) + " msg: " + r.Errmsg
}
