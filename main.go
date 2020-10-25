package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
				err := push(msg, sckey)
				if err != nil {
					log.Println(err)
					continue
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
	reg := regexp.MustCompile(`/mission/daily/redeem\?once=\d+`)
	once := reg.Find(b)
	return string(once), nil
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
	cookie = `__gads=ID=5750abdfb987fd64:T=1594472475:S=ALNI_MYOj1lha-PxWFAmp2MU5q-dQaVWYg; __cfduid=da8248aa276e1630fd177271c4b6e50d91602340711; A2="2|1:0|10:1603171047|2:A2|48:OGI2ODFiODgtOWFlOC00YTZhLWFiOWYtM2ZlM2I2YWQxNjk2|2fbb8a761ffc84ad63e86d760874eea36d12fd1b258186e2e9bb8593ecbce8af"; PB3_SESSION="2|1:0|10:1603453795|11:PB3_SESSION|36:djJleDoxNDkuMTI5LjkwLjg4OjQwNTM3NTMw|718e56e09d6f7ffe47c504bf33871e7f205d4d35223edbdd00327e00d7c47a4b"; V2EX_REFERRER="2|1:0|10:1603637475|13:V2EX_REFERRER|12:cmFpbmJvYXQ=|863302b0f8eb8d29bd6bb0f11aa25be8df52e74e31de852356d9582454ac8684"; A2O="2|1:0|10:1603642749|3:A2O|48:OGI2ODFiODgtOWFlOC00YTZhLWFiOWYtM2ZlM2I2YWQxNjk2|24fdcaffd398ee3c54348bfd7c397b9cf007bd9505b5fb681528b63358b2b17c"; V2EX_TAB="2|1:0|10:1603642749|8:V2EX_TAB|4:YWxs|91db34b0fa815da4073d1663a23ce40f591e2b47c91a48ee192ed359e31c3c2e"; V2EX_LANG=zhcn`
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
	reqs.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
	reqs.Header.Set("Cookie", cookie)
	rep, err := c.Do(reqs)
	if rep != nil {
		defer rep.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("httpget: %w", err)
	}
	b, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		return nil, fmt.Errorf("httpget: %w", err)
	}
	return b, nil
}

func getbalance() (int, error) {
	b, err := httpget(`https://www.v2ex.com/balance`)
	if err != nil {
		return 0, fmt.Errorf("getbalance: %w", err)
	}
	reg := regexp.MustCompile(`的每日登录奖励 [0-9]{1,4} 铜币`)
	msg := reg.Find(b)
	reg = regexp.MustCompile(`[0-9]{1,4}`)
	balance := reg.Find(msg)
	i, err := strconv.ParseInt(string(balance), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("getbalance: %w", err)
	}
	return int(i), nil
}

func push(msg, key string) error {
	msg = `text=v2ex签到信息&desp=` + url.QueryEscape(msg)
	reqs, err := http.NewRequest("POST", "https://sc.ftqq.com/"+key+".send", strings.NewReader(msg))
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	reqs.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rep, err := c.Do(reqs)
	if rep.Body != nil {
		defer rep.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	b, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	e := returnmsg{}
	err = json.Unmarshal(b, &e)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	if e.Errno != 0 {
		return fmt.Errorf("push: %w", errors.New(e.Errmsg))
	}
	return nil
}

//{"errno":0,"errmsg":"success","dataset":"done"}

type returnmsg struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
}
