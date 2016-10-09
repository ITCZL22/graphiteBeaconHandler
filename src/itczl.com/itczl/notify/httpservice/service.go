//Copyright 201606 itczl. All rights reserved.

//httpservice
package httpservice

// Grapihte-beacon request example:
// alert=UVE%3Amain_feed_qps&desc=%5BUVE+Alert%5D+WARNING+%3CUVE%3Amain_feed_qps%3E+%28itczl.access.qps.itczl-service-main_feed%29+failed.+Current+value%3A+9.8K&graph_url=http%3A%2F%2Fgraphite.itczl.pw%2Frender%2F%3Ftarget%3Ditczl.access.qps.itczl-service-main_feed%26from%3D-10minute%26until%3D-0second&level=warning&rule=warning%3A+%3E+2000&target=itczl.access.qps.itczl-service-main_feed&value=9839.75
//
// alert=UVE:main_feed_qps
// desc=[UVE Alert] WARNING <UVE:main_feed_qps> (itczl.access.qps.itczl-service-main_feed) failed. Current value: 9.8K
// graph_url=http://graphite.itczl.com/render/?target=itczl.access.qps.itczl-service-main_feed
// from=-10minute
// until=-0second
// level=warning
// rule=warning: > 2000
// target=itczl.access.qps.itczl-service-main_feed
// value=9839.75

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"itczl.com/itczl/notify/config"
)

type serverContext struct {
	sc                            config.Config
	to                            []string
	notice                        map[string]bool
	from                          string
	req                           *http.Request
	webhook, username, icon_emoji string
	level                         string
	reqc                          int
	tr                            *http.Transport
	client                        *http.Client
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

	log.Printf("Listen: %s\n", c.sc["hosts"].Host)
	err := s.ListenAndServe()
	if err != nil {
		return fmt.Errorf("httpservice.Run error: %v\n", err)
	}
	return nil
}

func (c *serverContext) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	req.ParseForm()
	allQuery := req.Form.Encode()

	c.reqc++
	status := http.StatusOK
	var buf bytes.Buffer
	var alert string
	var level string
	var noemail string
	var desc string
	var v config.AlertConf
	var ok bool

	if req.Form["alert"] == nil {
		status = http.StatusBadRequest
		buf.WriteString("`alert` parameter is required")
		goto RESP
	}

	alert = req.Form["alert"][0]
	if strings.ContainsAny(alert, ":") {
		alert = strings.SplitN(alert, ":", 2)[0]
	}

	v, ok = c.sc[alert]
	if !ok {
		status = http.StatusBadRequest
		buf.WriteString(fmt.Sprintf("Configuration NOT found for `alert`=%s", alert))
		goto RESP
	}

	c.to = v.MailTo
	c.from = v.MailFrom
	c.webhook = v.Slack.Webhook
	c.username = v.Slack.Username
	c.notice = v.Notice
	c.req = req

	if req.Form["level"] == nil {
		status = http.StatusBadRequest
		buf.WriteString("`level` parameter is required")
		goto RESP
	}

	c.level = req.Form["level"][0]

	if req.Form["desc"] == nil {
		status = http.StatusBadRequest
		buf.WriteString("`desc` parameter is required")
		goto RESP
	}

	desc = req.Form["desc"][0]

	if req.Form["noemail"] != nil && len(req.Form["noemail"]) > 0 {
		noemail = req.Form["noemail"][0]
	}

	if noemail != "1" && c.notice["mail"] {
		if err := c.mail(); err != nil {
			status = http.StatusInternalServerError
			log.Printf("Mail Error: %v\n", err)
			buf.WriteString(fmt.Sprintf("Email Error: %v\n", err))
		} else {
			buf.WriteString("Email OK\n")
		}
	}

	if c.notice["slack"] {
		log.Printf("Slack Web Hook: %s\n", c.webhook)
		if err := c.slack(); err != nil {
			status = http.StatusBadGateway
			log.Printf("Slack Error: %v\n", err)
			buf.WriteString(fmt.Sprintf("%v\n", err))
		} else {
			buf.WriteString("Slack OK\n")
		}
	}

RESP:
	w.WriteHeader(status)
	w.Write(buf.Bytes())
	log.Printf("Access [%d] [%d] [%s] level: %s, noemail: %s, desc: %s, Path: %s, Query: %s\n", status, c.reqc, req.Method, level, noemail, desc, req.URL.Path, allQuery)
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

type SlackPayload struct {
	IconEmoji   string            `json:"icon_emoji"`
	Username    string            `json:"username"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type SlackAttachment struct {
	Title    string       `json:"title"`
	Pretext  string       `json:"pretext"`
	Text     string       `json:"text"`
	Color    string       `json:"color"`
	Fallback string       `json:"fallback"`
	Fields   []SlackField `json:"fields"`
}

func (c *serverContext) slack() error {
	globalText := ""
	emoji := ""
	color := ""
	title := c.req.Form["alert"][0]
	text := c.req.Form["desc"][0]
	pretext := ""
	fallback := text

	if c.req.Form["title"] != nil { //Override the `alert` field, if presents
		title = c.req.Form["title"][0]
	}

	if c.req.Form["graph_url"] != nil { //Grapihte-beacon compatible
		text += "\n<" + c.req.Form["graph_url"][0] + "|Graphite Link>"
	}

	if c.req.Form["pretext"] != nil { //Override the `text` field if presents
		pretext = c.req.Form["pretext"][0]
	}

	if c.req.Form["globaltext"] != nil {
		globalText = c.req.Form["globaltext"][0]
	}

	if c.req.Form["fallback"] != nil {
		fallback = c.req.Form["fallback"][0]
	}

	if c.level == "warning" {
		emoji = ":warning:"
		color = "warning"
	} else if c.level == "critical" {
		emoji = ":x:"
		color = "danger"
	} else { //normal
		emoji = ":white_check_mark:"
		color = "good"
	}

	item := SlackAttachment{
		Title:    title,
		Text:     text,
		Pretext:  pretext,
		Color:    color,
		Fallback: fallback,
	}

	if c.req.Form["target"] != nil && c.req.Form["rule"] != nil && c.req.Form["value"] != nil { //Grapihte-beacon compatible
		//text += c.req.Form["target"][0] + " " + c.req.Form["value"][0] + " " + c.req.Form["rule"][0]
		item.Fields = append(item.Fields, SlackField{"Target", c.req.Form["target"][0], false})
		item.Fields = append(item.Fields, SlackField{"Value", c.req.Form["value"][0], false})
		item.Fields = append(item.Fields, SlackField{"Rule", c.req.Form["rule"][0], false})
	}

	payload := SlackPayload{}
	payload.Text = globalText
	payload.IconEmoji = emoji
	payload.Username = c.username
	payload.Attachments = append(payload.Attachments, item)

	p, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("httpservice.slack marshal error: %v\n", err)
	}

	if c.client == nil {
		log.Printf("Initializing Client...")
		if c.tr == nil {
			c.tr = &http.Transport{
				DisableKeepAlives: false,
				//MaxIdleConns:        10,
				MaxIdleConnsPerHost: 2,
				//IdleConnTimeout:     60 * time.Second,
			}
		}
		c.client = &http.Client{
			Transport: c.tr,
			//Timeout: time.Second * 5,
		}
	}
	resp, err := c.client.PostForm(c.webhook, url.Values{"payload": {string(p)}})
	if err != nil {
		return fmt.Errorf("httpservice.slack request slack error: %v\n", err)
	}
	log.Printf("Slack HTTP status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Slack Response: %s\n", string(body))
		return fmt.Errorf("Slack error: %s", resp.Status)
	}
	return nil
}
