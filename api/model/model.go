package model

import (
	"errors"
	"fmt"
	"github.com/crain-cn/event-mesh/cmd/config"
	event_versioned "github.com/crain-cn/event-mesh/pkg/k8s/client/clientset/versioned"
	"github.com/crain-cn/event-mesh/pkg/logging"
	monitoring_versioned "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/client-go/rest"
	"time"
)

var (
	Db      *gorm.DB
	DbPlat  *gorm.DB
	Clients *ClientManager
	log     = logging.DefaultLogger.WithField("component", "db")
)

type ClientManager struct {
	client  *event_versioned.Clientset
	mclient *monitoring_versioned.Clientset
}

func NewClientManager(config *rest.Config) *ClientManager {
	client, err := event_versioned.NewForConfig(config)
	if err != nil {
		//return nil, errors.Wrap(err, "instantiating kubernetes client failed")
	}

	mclient, err := monitoring_versioned.NewForConfig(config)
	if err != nil {
		//	return nil, errors.Wrap(err, "instantiating monitoring client failed")
	}
	Clients = &ClientManager{
		client:  client,
		mclient: mclient,
	}
	return Clients
}

type MysqlConn struct {
	Db           *gorm.DB
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
}

func NewMysqlConn(
	mysqlOpt *config.MysqlOpt,
	mysqlPlatformOpt *config.MysqlOpt,
	maxIdleConns int,
	maxOpenConns int,
	maxLifetime time.Duration) error {
	mysqlConn := &MysqlConn{
		MaxIdleConns: maxIdleConns,
		MaxOpenConns: maxOpenConns,
		MaxLifetime:  maxLifetime}
	err := mysqlConn.Setup(mysqlOpt)
	if err != nil {
		log.Error(err)
	}

	err = mysqlConn.SetupPlatform(mysqlPlatformOpt)
	if err != nil {
		log.Error(err)
	}

	return err
}

func (m *MysqlConn) SetupPlatform(opt *config.MysqlOpt) error {
	if opt == nil {
		return errors.New("unable get mysqlopt ")
	}
	var dialector gorm.Dialector
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local", opt.Username, opt.Password, opt.Host, opt.Port, opt.Database)
	dialector = mysql.New(mysql.Config{
		DSN:                       dsn,   // data source name
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	})
	var err error
	DbPlat, err = gorm.Open(dialector, &gorm.Config{
		Logger:                                   logging.NewGormLogger(),
		DisableForeignKeyConstraintWhenMigrating: false,
	})

	if err != nil {
		log.Error(err)
	}

	sqlDB, err := DbPlat.DB()

	if err != nil {
		log.Error("connect db server failed.")
	}
	sqlDB.SetMaxIdleConns(m.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Second * 600)
	if err := sqlDB.Ping(); err != nil {
		log.Error(err)
	}
	return nil
}

func (m *MysqlConn) Setup(opt *config.MysqlOpt) error {
	if opt == nil {
		return errors.New("unable get mysqlopt ")
	}
	var dialector gorm.Dialector
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local", opt.Username, opt.Password, opt.Host, opt.Port, opt.Database)
	dialector = mysql.New(mysql.Config{
		DSN:                       dsn,   // data source name
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	})
	var err error
	Db, err = gorm.Open(dialector, &gorm.Config{
		Logger:                                   logging.NewGormLogger(),
		DisableForeignKeyConstraintWhenMigrating: false,
	})

	if err != nil {
		log.Error(err)
	}

	sqlDB, err := Db.DB()

	if err != nil {
		log.Error("connect db server failed.")
	}
	sqlDB.SetMaxIdleConns(m.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Second * 600)
	if err := sqlDB.Ping(); err != nil {
		log.Error(err)
	}
	/*
		Db.Set("gorm:table_options", "ENGINE=InnoDB").
			AutoMigrate(
				&AlertHistory{},
				&EventHistory{},
				&AlertRule{},
				&RuleExpr{},
				&EventRule{},
				&Receiver{},
				&ReceiverDog{},
				&ReceiverWebhook{},
				&ReceiverYach{},
			)*/
	return nil
}
