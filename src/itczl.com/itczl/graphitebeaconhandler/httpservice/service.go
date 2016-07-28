//Copyright 201606 itczl. All rights reserved.

//httpservice
package httpservice

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"itczl.com/itczl/graphitebeaconhandler/config"
)

type serverContext struct {
	sc                            config.Config
	to                            []string
	notice                        map[string]bool
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

	name := req.Form["alert"][0]
	if strings.ContainsAny(name, ":") {
		name = strings.SplitN(name, ":", 2)[0]
	}
	v := c.sc[name]
	c.to = v.MailTo
	c.from = v.MailFrom
	c.webhook = v.Slack.Webhook
	c.username = v.Slack.Username
	c.notice = v.Notice
	c.req = req

	if c.req.Form["level"][0] == "warning" {
		c.icon_emoji = ":warning:"
	} else if c.req.Form["level"][0] == "normal" {
		c.icon_emoji = ":white_check_mark:"
	} else if c.req.Form["level"][0] == "critical" {
		c.icon_emoji = ":x:"
	}

	if c.req.Form["noemail"] == nil && c.notice["mail"] {
		if err := c.mail(); err != nil {
			log.Printf("%v\n", err)
			io.WriteString(w, "send mail error\n")
		}
	}

	if c.notice["slack"] {
		if err := c.slack(); err != nil {
			log.Printf("%v\n", err)
			io.WriteString(w, "invoke slack error\n")
		}
	}
}

func (c *serverContext) mail() error {
	msg := []byte("To: " + strings.Join(c.to, " ") + "\r\n" +
		"Subject: " + c.username + "\r\n" +
		"\r\n" +
		c.req.Form["desc"][0])
	err := smtp.SendMail("127.0.0.1:25", nil, c.from, c.to, msg)
	if err != nil {
		return fmt.Errorf("httpservice.mail send mail error: %v\n", err)
	}
	return nil
}

func (c *serverContext) slack() error {
	type postData struct {
		Text       string `json:"text"`
		Username   string `json:"username"`
		Icon_emoji string `json:"icon_emoji"`
	}
	pd := postData{c.req.Form["desc"][0], c.username, c.icon_emoji}
	p, err := json.Marshal(pd)
	if err != nil {
		return fmt.Errorf("httpservice.slack marshal error: %v\n", err)
	}

	resp, err := http.PostForm(c.webhook, url.Values{"payload": {string(p)}})
	if err != nil {
		return fmt.Errorf("httpservice.slack request slack error: %v\n", err)
	}
	fmt.Println(resp.Status)
	return nil
}
