package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
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
			_, err = httpget(`https://www.v2ex.com/` + once)
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
	}
	Push("签到失败", chatID, 5)
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
		Push("cookie 失效", chatID, 5)
		panic("cookie 失效")
	}
	if bytes.Contains(b, []byte(`每日登录奖励已领取`)) {
		return false, nil
	}
	return true, nil
}

var (
	c           = http.Client{Timeout: 5 * time.Second}
	cookie      string
	telegramkey string
	chatID      string
)

func init() {
	cookie = os.Getenv("v2exCookie")
	if cookie == "" {
		Push("你 cookie 呢？", chatID, 5)
		panic("你 cookie 呢？")
	}
	telegramkey = os.Getenv("telegramkey")
	chatID = os.Getenv("chatID")
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

func push(message, chatID string) error {
	message = url.QueryEscape(message)
	msg := "chat_id=" + chatID + "&text=" + message
	req, err := http.NewRequest("POST", telegramkey+"/sendMessage", strings.NewReader(msg))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reps, err := c.Do(req)
	if reps != nil {
		defer reps.Body.Close()
	}
	if err != nil {
		return err
	}
	t, err := ioutil.ReadAll(reps.Body)
	if err != nil {
		return err
	}
	var ok isok
	json.Unmarshal(t, &ok)
	if !ok.OK {
		return Pusherr
	}
	return nil
}

var Pusherr = errors.New("推送失败")

type isok struct {
	OK bool `json:"ok"`
}

func Push(message, chatID string, i int) {
	if telegramkey == "" {
		return
	}
	if i <= 0 {
		log.Println("推送失败", message)
		return
	}
	err := push(message, chatID)
	if err != nil {
		log.Println(err, message)
		time.Sleep(1 * time.Second)
		Push(message, chatID, i-1)
	}
}
