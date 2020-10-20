package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
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
			if !cansign {
				log.Println("签到失败，尝试重签")
				continue
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
		return "", err
	}
	reg := regexp.MustCompile(`/mission/daily/redeem\?once=\d+`)
	once := reg.Find(b)
	return string(once), nil
}

func check() (bool, error) {
	b, err := httpget("https://www.v2ex.com/mission/daily")
	if err != nil {
		return false, err
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
)

func init() {
	cookie = os.Getenv("v2exCookie")
	cookie = `__gads=ID=5750abdfb987fd64:T=1594472475:S=ALNI_MYOj1lha-PxWFAmp2MU5q-dQaVWYg; A2="2|1:0|10:1602252951|2:A2|48:YjczY2JiY2YtNDZkZC00YTA4LWJiMTMtNTg0OTRlMTZlMTkx|d84f488a3ff39966ce9eb0937beb9bd86ee3970eb98c6f18dd7af3a958c92ff1"; __cfduid=da8248aa276e1630fd177271c4b6e50d91602340711; PB3_SESSION="2|1:0|10:1603012645|11:PB3_SESSION|36:djJleDoxNDkuMTI5LjkwLjg4OjMwMDU1NjM2|bc9c39b9abdb39fb7ce32ff213a8949b492d85d96a596d60f3c37bbe56962b8b"; V2EX_REFERRER="2|1:0|10:1603037291|13:V2EX_REFERRER|12:anN5emNoZW4=|5d4f87f505e261fb663a99cf2037d95e4951fcb1647c37befefce6d727ec2690"; V2EX_LANG=zhcn; A2O="2|1:0|10:1603170566|3:A2O|48:YjczY2JiY2YtNDZkZC00YTA4LWJiMTMtNTg0OTRlMTZlMTkx|7c9dfc10216f48a640ffc27cd3b94a4b17e61e8a4a28211f37b520ec6ae13fb9"; V2EX_TAB="2|1:0|10:1603170566|8:V2EX_TAB|4:YWxs|6ba058c777d135ef227b6c0f7f417f022af312f8941ab55c1cbd417ffac88ad9"`
	if cookie == "" {
		panic("你 cookie 呢？")
	}
}

func httpget(url string) ([]byte, error) {
	reqs, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	reqs.Header.Set("Accept", "*/*")
	reqs.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
	reqs.Header.Set("Cookie", cookie)
	rep, err := c.Do(reqs)
	if rep != nil {
		defer rep.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(rep.Body)
}
