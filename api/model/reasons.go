package model

type EventReasons struct {
	ID      uint   `json:"id" gorm:"primarykey"`
	Name    string `json:"name" example:"normal event notify"`
	Label   string `json:"label" example:"app"`
	ClassId int    `json:"class_id"`
	Doc     string `json:"doc" example:"On"`
}

func (a *EventReasons) TableName() string {
	return "notification_reasons"
}

func GetEventReasonsAll() map[string]string {
	var eventReasons []*EventReasons
	tx := Db.Table("notification_reasons")
	tx.Order("id desc").Limit(1000).Find(&eventReasons)
	reasons := make(map[string]string,0)
	for _,value := range eventReasons {
		reasons[value.Name] = value.Label
	}
	return reasons
}
