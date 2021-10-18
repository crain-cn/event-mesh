package yach

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/crain-cn/event-mesh/api/model"
	"github.com/crain-cn/event-mesh/pkg/config"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"github.com/crain-cn/event-mesh/pkg/notify"
	"github.com/crain-cn/event-mesh/pkg/template"
	"github.com/prometheus/alertmanager/types"
	commoncfg "github.com/prometheus/common/config"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

//}
const (
	TextType     = "text"
	MarkdownType = "markdown"
)

var yachURL = "https://yach-oapi.zhiyinlou.com/robot/send"

type YachResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Notifier implements a Notifier for Slack notifications.
type Notifier struct {
	conf    *config.YachConfig
	tmpl    *template.Template
	client  *http.Client
	retrier *notify.Retrier
	logger  *logrus.Entry
}

type YachTextMsg struct {
	Content string `json:"content"`
}

type YachMarkdownMsg struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type YachAt struct {
	AtMobiles []string `json:"atMobiles"`
	AtYachIds []string `json:"atWorkCodes"`
	IsAtAll   bool     `json:"isAtAll"`
}

type Message struct {
	Msgtype  string           `json:"msgtype"`
	Text     *YachTextMsg     `json:"text,omitempty"`
	Markdown *YachMarkdownMsg `json:"markdown,omitempty"`
	At       *YachAt          `json:"at"`
}

// New returns a new Slack notification handler.
func New(conf *config.YachConfig, t *template.Template) (*Notifier, error) {
	client, err := commoncfg.NewClientFromConfig(*conf.HTTPConfig, "yach", false)
	if err != nil {
		return nil, err
	}

	return &Notifier{
		conf:    conf,
		tmpl:    t,
		client:  client,
		logger:  logging.DefaultLogger.WithField("notify", "yach"),
		retrier: &notify.Retrier{},
	}, nil
}

//TODO Add Keyword support
func (n *Notifier) sign() (v url.Values) {
	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	hmacHash := hmac.New(sha256.New, []byte(n.conf.Secret))
	hmacHash.Write([]byte(timestamp + "\n" + n.conf.Secret))
	r := hmacHash.Sum(nil)
	sign := base64.StdEncoding.EncodeToString(r)
	v = url.Values{}
	v.Add("timestamp", timestamp)
	v.Add("sign", sign)
	v.Add("access_token", n.conf.AccessToken)
	return v
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	key, err := notify.ExtractGroupKey(ctx)
	if err != nil {
		return false, err
	}
	//n.logger.Info(key)
	//data := notify.GetTemplateData(ctx, n.tmpl, as, n.logger)

	//tmpl := notify.TmplText(n.tmpl, data, &err)
	if err != nil {
		return false, err
	}
	//n.logger.Info(tmpl(n.conf.Message))

	title := fmt.Sprintf("容器告警")
	//text := n.genMarkdown(as)
	msg := n.genMarkdown(title,as)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return false, err
	}

	v := n.sign()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s?%s", yachURL, v.Encode()), &buf)
	if err != nil {
		return true, err
	}
	resp, err := n.client.Do(req.WithContext(ctx))
	if err != nil {
		return true, notify.RedactURL(err)
	}
	defer notify.Drain(resp)

	if resp.StatusCode != 200 {
		return true, fmt.Errorf("unexpected status code %v", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		n.logger.WithFields(logrus.Fields{"response": string(respBody), "iincident": key}).WithError(err).Error()
		return true, err
	}
	yachResponse := YachResponse{}
	err = json.Unmarshal(respBody, &yachResponse)
	if yachResponse.Code != 200 {

	}
	n.logger.WithFields(logrus.Fields{"response": string(respBody), "iincident": key}).Debug()
	defer notify.Drain(resp)

	return true, nil
}

func (n *Notifier) genMarkdown(title string,as []*types.Alert) *Message {

	msg := &Message{
		Msgtype: MarkdownType,
		Markdown: &YachMarkdownMsg{
			Title: title,
		},
	}

	msgList := make([]string, 0)
	envStr := os.Getenv("env")
	for key, alert := range as {
		labels := &model.AlertLables{}
		isEvent := false
		for key, v := range alert.Labels {
			switch key {
			case "alertname":
				labels.Alertname = string(v)
			case "severity":
				labels.Severity = string(v)
			case "pod":
				labels.Pod = string(v)
			case "namespace":
				labels.Namespace = string(v)
			case "node":
				labels.Node = string(v)
			case "cluster":
				labels.Cluster = string(v)
			}
		}

		if _, ok := alert.Labels["event_reason"]; ok {
			isEvent = true
			msgList = append(msgList, fmt.Sprintf("\n## 『事件通知』"))
			msgList = append(msgList, fmt.Sprintf("\n### 事件%d: %s", key, alert.Labels["event_reason"]))

		} else {
			if !alert.EndsAt.IsZero() {
				msgList = append(msgList, fmt.Sprintf("\n## 『告警恢复』"))
			} else {
				msgList = append(msgList, fmt.Sprintf("\n## 『告警通知』"))
			}
			msgList = append(msgList, fmt.Sprintf("\n### 指标%d: %s", key, labels.Alertname))
		}

		if _, ok := alert.Labels["cn_reason"]; ok   {
			if alert.Labels["cn_reason"] != "" {
				msgList = append(msgList, fmt.Sprintf("\n### 原因: %s", alert.Labels["cn_reason"]))
			}
		}


		msgList = append(msgList, fmt.Sprintf("### 级别: %s", labels.Severity))
		msgList = append(msgList, fmt.Sprintf("### 组件: %s", alert.Labels["obj_kind"]))
		message := string(alert.Annotations["message"])
		if len(message) > 0 {
			msgList = append(msgList, fmt.Sprintf("### 描述:"))
			msgList = append(msgList, fmt.Sprintf("> %s", message))
		}
		msgList = append(msgList, fmt.Sprintf("### Lables:"))
		msgList = append(msgList, fmt.Sprintf("> cluster：%s", labels.Cluster))
		if len(labels.Node) > 0 {
			msgList = append(msgList, fmt.Sprintf("> node: %s", labels.Node))
		} else {
			if _, ok := alert.Labels["source_host"]; ok   {
				if alert.Labels["source_host"] != "" && alert.Labels["source_host"] != "eci" {
					msgList = append(msgList, fmt.Sprintf("> node: %s", alert.Labels["source_host"]))
				}
			}
		}

		if len(labels.Namespace) > 0 {
			msgList = append(msgList, fmt.Sprintf("> namespace: %s", labels.Namespace))
		}
		if len(labels.Pod) > 0 {
			msgList = append(msgList, fmt.Sprintf("> pod: %s", labels.Pod))
		}

		if alert.Labels["obj_name"] != "" {
			msgList = append(msgList, fmt.Sprintf("> obj_name: %s", alert.Labels["obj_name"]))
		}

		if len(labels.Deployment) > 0 {
			msgList = append(msgList, fmt.Sprintf("> deployment: %s", labels.Deployment))
		}
		time := string(alert.StartsAt.Format("2006-01-02 15:04:05"))
		msgList = append(msgList, fmt.Sprintf("\n### 开始时间: %s", time))

		if !isEvent && !alert.EndsAt.IsZero() {
			time2 := string(alert.EndsAt.Format("2006-01-02 15:04:05"))
			msgList = append(msgList, fmt.Sprintf("\n### 恢复时间: %s", time2))
		}
		urlStr := "https://cloud.tal.com/"
		if envStr == "test" || envStr == "dev" {
			urlStr = "https://cloud-test.tal.com/"
		}
		timeStr := url.QueryEscape(time)
		params := fmt.Sprintf("reason=%s&severity=%s&cluster=%s&namespace=%s&obj_kind=%s&source_host=%s&datetime=%s",alert.Labels["event_reason"],strings.ToLower(labels.Severity),labels.Cluster,labels.Namespace,alert.Labels["obj_kind"],labels.Node,timeStr)
		msgList = append(msgList, fmt.Sprintf("\n### 详情链接: %s", fmt.Sprintf("%shunter/notification/events?%s",urlStr,params)))

		if _, ok := alert.Labels["workcode"]; ok   {
			if alert.Labels["workcode"] != "" {
				msg.At = &YachAt{
					AtMobiles: []string{},
					AtYachIds: []string{string(alert.Labels["workcode"])},
					IsAtAll: false,
				}
			}
		}
	}

	msgList = append(msgList, "\n#### 容器iaas团队为您稳定性保驾护航！")
	msg.Markdown.Text =  strings.Join(msgList, "\n")



	return msg
}
