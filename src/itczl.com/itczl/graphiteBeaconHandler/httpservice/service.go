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

	"uve.io/uve/graphiteBeaconHandler/config"
)

type serverContext struct {
	sc *config.Config
}

func Run(conf *config.Config) error {
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
	//获取req参数
	req.ParseForm()

	var to []string
	var username string
	for _, v := range *c.sc {
		if strings.Contains(req.Form["name"][0], v.Name) {
			to = v.Mail
			username = v.Slack.Username
		}
	}

	if err := mail(to, req.Form["level"][0], req.Form["value"][0], username); err != nil {
		log.Printf("%v\n", err)
		io.WriteString(w, "send mail error")
		return
	}

	if err := slack(req.Form["level"][0], req.Form["value"][0], username); err != nil {
		log.Printf("%v\n", err)
		io.WriteString(w, "invoke slack error")
		return
	}

	io.WriteString(w, "ok\n")
}

func mail(to []string, level, value, username string) error {
	msg := []byte("To: zhenliang@staff.weibo.com\r\n" +
		"Subject: test!\r\n" +
		"\r\n" +
		username + " " + level + " " + value)
	err := smtp.SendMail("127.0.0.1:25", nil, "uve-graphite-beacon@56.uve.mobile.sina.cn", to, msg)
	if err != nil {
		return fmt.Errorf("httpservice.mail send mail error: %v\n", err)
	}
	return nil
}

func slack(level, value, username string) error {
	postData := "{\"text\":\"" + value + "\",\"channel\":\"" + level + "\",\"icon_emoji\":\"ghost\",\"username\":\"" + username + "\"}"

	resp, err := http.PostForm("https://hooks.slack.com/services/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", url.Values{"payload": {postData}})
	if err != nil {
		return fmt.Errorf("httpservice.slack invoke slack error: %v\n", err)
	}
	fmt.Println(resp)
	return nil
}
