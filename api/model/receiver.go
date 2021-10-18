package model

import (
	"context"
	v1 "github.com/crain-cn/event-mesh/pkg/k8s/apis/notification/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type Receiver struct {
	ID            uint             `json:"id" gorm:"primarykey"`
	GroupRefer    uint             `json:"group_id" gorm:"index"`
	Name          string           `json:"name" gorm:"uniqueIndex"`
	Group         *AppGroup        `json:"group" example:"hudong" gorm:"foreignKey:GroupRefer"`
	Default       bool             `json:"default"`
	Type          string           `json:"type" example:"webhook"`
	WebhookConfig *ReceiverWebhook `json:"webhook_config"`
	DogConfig     *ReceiverDog     `json:"dog_config"`
	YachConfig    *ReceiverYach    `json:"yach_config"`
	CreatedAt     time.Time        `json:"created_at"`
}

func (a *Receiver) TableName() string {
	return "notification_receiver"
}

type ReceiverWebhook struct {
	ReceiverID uint   `json:"receiver_id"`
	Url        string `json:"url""`
}

func (a *ReceiverWebhook) TableName() string {
	return "notification_receiver_webhook"
}

type ReceiverDog struct {
	ReceiverID uint  `json:"receiver_id"`
	TaskId     int32 `json:"task_id,"`
}

func (a *ReceiverDog) TableName() string {
	return "notification_receiver_dog"
}

type ReceiverYach struct {
	ReceiverID  uint   `json:"receiver_id"`
	AccessToken string `json:"access_token"`
	Secret      string `json:"secret,"`
	Keyword     string `json:"keyword,"`
}

func (a *ReceiverYach) TableName() string {
	return "notification_receiver_yach"
}

type ReceiverListRepose struct {
	Code    int         `json:"code" example:"0"`
	Stat    int         `json:"stat" example:"0"`
	Message string      `json:"msg" example:""`
	Data    []*Receiver `json:"data"`
}

type ReceiverRepose struct {
	Code    int       `json:"code" example:"0"`
	Stat    int       `json:"stat" example:"0"`
	Message string    `json:"msg" example:""`
	Data    *Receiver `json:"data"`
}

var ReceiverSample = &Receiver{
	Name: "bot3", Default: true, Type: "yach",
	YachConfig: &ReceiverYach{AccessToken: "XXXXX", Secret: "S-XXXX", Keyword: "KEY"},
}

var ReceiversSample = []*Receiver{
	{Name: "bot1", Default: true, Type: "dog",
		DogConfig: &ReceiverDog{TaskId: 111},
	},
	{Name: "bot2", Default: true, Type: "webhook",
		WebhookConfig: &ReceiverWebhook{Url: "http://baidu.com"},
	},
	{Name: "bot3", Default: true, Type: "yach", YachConfig: &ReceiverYach{AccessToken: "XXXXX", Secret: "S-XXXX", Keyword: "KEY"}},
}

var receiverTypes = []string{
	"dog", "webhook", "yach",
}

func ListReceiver(input *Receiver) []*Receiver {
	var receivers []*Receiver
	var receiver Receiver
	tx := Db.Model(&receiver)
	for _, t := range receiverTypes {
		if input.Type == t {
			receiver.Type = input.Type
			break
		}
	}
	if len(input.Name) > 0 {
		tx = tx.Where("name like ?", "%"+input.Name+"%")
	}
	if input.GroupRefer > 0 {
		tx = tx.Where("group_refer =?", input.GroupRefer)
	}
	tx = tx.Limit(100).Order("id desc")
	tx.Find(&receivers)
	//stmt := tx.Session(&gorm.Session{DryRun: true}).Find(&receivers).Statement
	//log.Info(stmt.SQL.String())

	for key, receiver := range receivers {
		switch receiver.Type {
		case "dog":
			receiver.DogConfig = GetReceiverDog(receiver.ID)
		case "webhook":
			receiver.WebhookConfig = GetReceiverWebhook(receiver.ID)
		case "yach":
			receiver.YachConfig = GetReceiverYach(receiver.ID)
		}
		receivers[key] = receiver
	}
	return receivers
}

func GetReceiverYach(id uint) *ReceiverYach {
	receiverYach := &ReceiverYach{}
	Db.Where(&ReceiverYach{ReceiverID: id}).Limit(1).First(receiverYach)
	return receiverYach
}

func GetReceiverDog(id uint) *ReceiverDog {
	receiverDog := &ReceiverDog{}
	Db.Where(&ReceiverDog{ReceiverID: id}).Limit(1).First(receiverDog)
	return receiverDog
}

func GetReceiverWebhook(id uint) *ReceiverWebhook {
	receiverWebhook := &ReceiverWebhook{}
	Db.Where(&ReceiverWebhook{ReceiverID: id}).Limit(1).First(receiverWebhook)
	return receiverWebhook
}

func GetReceiver(id uint) *Receiver {
	receiver := &Receiver{}
	Db.Where(&Receiver{ID: id}).First(receiver)
	return receiver
}

func GetReceiverById(id uint) *Receiver {
	receiver := &Receiver{}
	Db.Where(&Receiver{ID: id}).First(receiver)
	return receiver
}

func UpdateReceiver(old, update *Receiver) (error, *Receiver) {
	delReceiverResource(old.Name)
	receiver := &Receiver{}
	switch update.Type {
	case "dog":
		update.DogConfig.ReceiverID = update.ID
		if old.Type != update.Type {
			Db.Model(&ReceiverDog{}).Create(update.DogConfig)
		} else {
			Db.Model(&ReceiverDog{}).Where(&ReceiverDog{ReceiverID: update.ID}).Updates(update.DogConfig)
		}
	case "webhook":
		update.WebhookConfig.ReceiverID = update.ID
		update.WebhookConfig.Url = strings.Trim(update.WebhookConfig.Url, " ")
		if old.Type != update.Type {
			Db.Model(&ReceiverWebhook{}).Create(update.WebhookConfig)
		} else {
			Db.Model(&ReceiverWebhook{}).Where(&ReceiverWebhook{ReceiverID: update.ID}).Updates(update.WebhookConfig)
		}
	case "yach":
		update.YachConfig.ReceiverID = update.ID
		update.YachConfig.AccessToken = strings.Trim(update.YachConfig.AccessToken, " ")
		update.YachConfig.Secret = strings.Trim(update.YachConfig.Secret, " ")
		if old.Type != update.Type {
			Db.Model(&ReceiverYach{}).Create(update.YachConfig)
		} else {
			update.YachConfig.ReceiverID = update.ID
			Db.Model(&ReceiverYach{}).Where(&ReceiverYach{ReceiverID: update.ID}).Updates(update.YachConfig)
		}
	}
	addReceiverResource(update)
	update.DogConfig = nil
	update.WebhookConfig = nil
	update.YachConfig = nil
	result := Db.Model(&receiver).Where(&Receiver{ID: update.ID}).Updates(update)
	if result.Error == nil && update.Default {
		Db.Model(&Receiver{}).Where("id != ?", update.ID).
			Where("group_refer =?", update.GroupRefer).
			Update("default", 0)
	}
	return result.Error, GetReceiver(update.ID)
}

func DeleteReceiver(id uint) error {
	r := GetReceiver(id)
	tx := Db.Where("id =?", id).Delete(&Receiver{})
	//	stmt := tx.Session(&gorm.Session{DryRun: true}).Where("id =?", id).Delete(&Receiver{}).Delete(&Receiver{}).Statement
	//log.Info(stmt.SQL.String())
	if tx.Error == nil {
		delReceiverResource(r.Name)
	}
	return tx.Error
}

func AddReceiver(r *Receiver) error {
	result := Db.Create(r)
	if result.Error == nil && r.Default {
		Db.Model(&Receiver{}).Where("id != ?", r.ID).
			Where("group_refer = ?", r.Group).
			Update("default", 0)
	}
	return result.Error
}

func addReceiverResource(r *Receiver) {
	webhookConfig := v1.WebhookConfig{}
	yachConfig := v1.YachConfig{}
	dogConfig := v1.DogConfig{}
	switch r.Type {
	case "dog":
		dogConfig = v1.DogConfig{
			TaskId: r.DogConfig.TaskId, MaxAlerts: 1,
		}
	case "yach":
		yachConfig = v1.YachConfig{
			AccessToken: r.YachConfig.AccessToken,
			Secret:      r.YachConfig.Secret,
		}

	case "webhook":
		webhookConfig = v1.WebhookConfig{
			URL: &r.WebhookConfig.Url,
		}
	}
	_, err := Clients.client.NotificationV1().Receivers().Create(context.TODO(), &v1.Receiver{
		ObjectMeta: metav1.ObjectMeta{
			Name:   r.Name,
			Labels: map[string]string{
				//	"group": string(r.GroupRefer),
			},
		},
		Spec: v1.ReceiverSpec{
			Group:          string(r.GroupRefer),
			DefaultMethond: r.Default,
			WebhookConfig:  &webhookConfig,
			DogConfig:      &dogConfig,
			YachConfig:     &yachConfig,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		log.Error(err)
	}
}

func delReceiverResource(name string) {
	Clients.client.NotificationV1().Receivers().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

