package model

import (
	"time"
)

type AppGroup struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Name       string    `json:"name"`
	Desc       string    `json:"desc"`
	CreateUser string    `json:"create_user"`
	UpdateAt   time.Time `json:"update_at"`
	CreateAt   time.Time `json:"create_at"`
}

type XesApp struct {
	Deployment string `json:"deployment"`
	Namespace  string `json:"namespace"`
	GroupId    uint   `json:"group_id"`
}

type XesDeploy struct {
	Namespace string `json:"namespace"`
	Deployment string `json:"deployment"`
	ManagerId  int `json:"manager_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Workcode   string `json:"workcode"`
}


func getNamespacesByGroup(group uint) []*XesApp {
	xesApps := []*XesApp{}
	tx := DbPlat.Table("k8s_platform.xes_cloud_app").
		Select("deployment, namespace").
		Where("group_id  = ?", group)

	//stmt := tx.Session(&gorm.Session{DryRun: true}).Scan(items).Statement
	//log.Info(stmt.SQL.String())
	tx.Scan(&xesApps)
	return xesApps
}

func getGroupByApp(namespace string, deployment string) *XesApp {
	xesApp := &XesApp{}
	tx := DbPlat.Table("k8s_platform.xes_cloud_app").
		Select("deployment, namespace,  group_id").
		Where("namespace  = ?", namespace).
		Where("deployment  = ?", deployment)
	//stmt := tx.Session(&gorm.Session{DryRun: true}).Scan(items).Statement
	//log.Info(stmt.SQL.String())
	tx.Scan(xesApp)
	return xesApp
}

func getAppGroup(group_id uint) *AppGroup {
	AppGroup := &AppGroup{}
	tx := DbPlat.Table("k8s_platform.xes_cloud_app_group").
		Select("id,  name").
		Where("id  = ?", group_id)
	tx.Scan(AppGroup)
	return AppGroup
}

func GetAppUserAll() map[string]string {
	xesDeploy := []*XesDeploy{}
	tx := DbPlat.Table("k8s_platform.xes_cloud_app").
		Select("k8s_platform.xes_cloud_app.namespace as namespace,k8s_platform.xes_cloud_app.deployment as deployment,k8s_platform.xes_cloud_app.manager_id as manager_id,k8s_platform.xes_cloud_user.name as name,k8s_platform.xes_cloud_user.email as email,k8s_platform.xes_cloud_user.workcode as workcode").
		Joins("left join k8s_platform.xes_cloud_user  on k8s_platform.xes_cloud_user.id = k8s_platform.xes_cloud_app.manager_id").
		Where("k8s_platform.xes_cloud_user.name is NOT NULL")
	tx.Scan(&xesDeploy)

	users := make(map[string]string,0)
	for _,value := range xesDeploy {
		users[value.Namespace + "|" + value.Deployment] = value.Workcode
	}
	return users
}
