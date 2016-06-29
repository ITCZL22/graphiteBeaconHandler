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
	"time"

	"itczl.com/itczl/graphiteBeaconHandler/config"
)

type serverContext struct {
	sc                            config.Config
	to                            []string
	from                          string
	req                           *http.Request
	webhook, username, icon_emoji string
}

func Run(conf config.Config) error {
	c := &serverContext{}
	c.sc = conf
	s := &http.Server{
		Addr:           c.sc["hosts"].Host,
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

	v := c.sc[req.Form["alert"][0]]
	c.to = v.MailTo
	c.from = v.MailFrom
	c.webhook = v.Slack.Webhook
	c.username = v.Slack.Username
	c.req = req

	if c.req.Form["level"][0] == "warning" {
		c.icon_emoji = ":warning:"
	} else if c.req.Form["level"][0] == "normal" {
		c.icon_emoji = ":white_check_mark:"
	} else if c.req.Form["level"][0] == "critical" {
		c.icon_emoji = ":x:"
	}

	if err := c.mail(); err != nil {
		log.Printf("%v\n", err)
		io.WriteString(w, "send mail error\n")
		return
	}

	if err := c.slack(); err != nil {
		log.Printf("%v\n", err)
		io.WriteString(w, "invoke slack error\n")
		return
	}
}

func (c *serverContext) mail() error {
	msg := []byte("To: zhenliang@staff.weibo.com\r\n" +
		"Subject: UVE Alter\r\n" +
		"\r\n" +
		c.req.Form["desc"][0])
	err := smtp.SendMail("127.0.0.1:25", nil, c.from, c.to, msg)
	if err != nil {
		return fmt.Errorf("httpservice.mail send mail error: %v\n", err)
	}
	return nil
}

func (c *serverContext) slack() error {
	postData := "{\"text\":\"" + c.req.Form["desc"][0] + "\", \"username\":\"" + c.username + "\", \"icon_emoji\":\"" + c.icon_emoji + "\"}"

	resp, err := http.PostForm(c.webhook, url.Values{"payload": {postData}})
	if err != nil {
		return fmt.Errorf("httpservice.slack invoke slack error: %v\n", err)
	}
	fmt.Println(resp.Status)
	return nil
}
