//Copyright 201606 itczl. All rights reserved.

//httpservice
package httpservice

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"itczl.com/itczl/graphiteBeaconHandler/config"
)

type serverContext struct {
	sc                                  config.Config
	to                                  []string
	value, level                        string
	name, promote                       string
	icon_emoji, webhook, username, note string
}

func Run(conf config.Config) error {
	c := &serverContext{}
	c.sc = conf
	s := &http.Server{
		Addr:           "10.77.96.56:8883",
		Handler:        c,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := s.ListenAndServe()
	if err != nil {
		return fmt.Errorf("httpservice.Run error: %v\n", err)
	}
	return nil
}

func (c *serverContext) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//解析req参数
	req.ParseForm()

	v := c.sc[req.Form["name"][0]]
	c.to = v.Mail
	c.webhook = v.Slack.Webhook
	c.promote = v.Slack.Promote
	c.username = v.Slack.Username
	c.name = req.Form["name"][0]
	c.value = req.Form["value"][0]
	c.level = req.Form["level"][0]

	if c.level == "warning" {
		c.icon_emoji = ":warning:"
		c.note = " failed. "
	} else if c.level == "normal" {
		c.icon_emoji = ":white_check_mark:"
		c.note = " is back to normal. "
	} else if c.level == "critical" {
		c.icon_emoji = ":x:"
		c.note = " failed. "
	}

	if err := c.mail(); err != nil {
		log.Printf("%v\n", err)
		io.WriteString(w, "send mail error")
		return
	}

	if err := c.slack(); err != nil {
		log.Printf("%v\n", err)
		io.WriteString(w, "invoke slack error")
		return
	}
}

func (c *serverContext) mail() error {
	msg := []byte("To: zhenliang@staff.weibo.com\r\n" +
		"Subject: UVE Alter\r\n" +
		"\r\n" +
		strings.ToUpper(c.level) + " " + "[service:" + c.name + "] - " + c.promote + c.note + "\n\n\t\tValue: " + c.value)
	err := smtp.SendMail("127.0.0.1:25", nil, "uve-graphite-beacon@56.uve.mobile.sina.cn", c.to, msg)
	if err != nil {
		return fmt.Errorf("httpservice.mail send mail error: %v\n", err)
	}
	return nil
}

func (c *serverContext) slack() error {
	postData := "{\"text\":\"[" + c.username + "] " + strings.ToUpper(c.level) + " <service:" + c.name + ">" + c.promote + c.note + "Current value:" + c.value + "\", \"icon_emoji\":\"" + c.icon_emoji + "\", \"username\":\"" + c.username + "\"}"
	resp, err := http.PostForm(c.webhook, url.Values{"payload": {postData}})
	if err != nil {
		return fmt.Errorf("httpservice.slack invoke slack error: %v\n", err)
	}
	fmt.Println(resp)
	return nil
}
