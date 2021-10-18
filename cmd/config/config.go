package config

import (
	"errors"
	"github.com/crain-cn/event-mesh/pkg/logging"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var (
	log = logging.DefaultLogger.WithField("component", "db")
)

const DefaultEnv = "test"

type ConfigResolver struct {
	Databases  []*Database     `yaml:"databases"`
	Sessions   []*SessionStore `yaml:"sessions"`
	EventSinks []*EventSinks   `yaml:"eventSinks"`
}

type Database struct {
	Name   string    `yaml:"name"`
	Mysql  *MysqlOpt `yaml:"mysql"`
	Labels *Labels   `yaml:"labels"`
}

type SessionStore struct {
	Store  string    `yaml:"store"`
	Redis  *RedisOpt `yaml:"redis"`
	Labels *Labels   `yaml:"labels"`
}

type MysqlOpt struct {
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	Host     string `yaml:"host,omitempty" json:"host,omitempty"`
	Port     string `yaml:"port,omitempty" json:"port,omitempty"`
	Database string `yaml:"database,omitempty" json:"database,omitempty"`
}

type RedisOpt struct {
	Address     string `yaml:"address"`
	Password    string `yaml:"password"`
	PoolSise    int    `yaml:"pool_size"`
	ReadTimeout int    `yaml:"read_timeout"`
	IdleTimeout int    `yaml:"idle_timeout"`
}

type EventSinks struct {
	Cluster    string   `yaml:"cluster"`
	Filters    []string `yaml:"filters"`
	NotFilters []string `yaml:"notFilters"`
	SlsOpt     *SlsOpt  `yaml:"slsSink"`
}

type SlsOpt struct {
	Project         string `yaml:"project"`
	RegionId        string `yaml:"regionId"`
	AccessKeyId     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
}

type Labels struct {
	Env string `yaml:"env"`
}

func NewResolver(file string) (*ConfigResolver, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("config file not IsNotExist", err)
			return nil, err
		}
		return nil, err
	}
	config := &ConfigResolver{}
	err = yaml.UnmarshalStrict(b, config)
	if err != nil {
		log.Error("NewResolver", err)
	}
	return config, nil
}

func (c *ConfigResolver) GetDbConfig(dbname string, env string) (*MysqlOpt, error) {
	if len(env) == 0 {
		env = DefaultEnv
	}
	for _, db := range c.Databases {
		if db.Labels.Env != env {
			continue
		}
		if db.Name == dbname {
			return db.Mysql, nil
		}
	}
	return nil, errors.New("unknown mysql")
}

func (c *ConfigResolver) GetSessionStore(store string, env string) (*SessionStore, error) {
	if len(env) == 0 {
		env = DefaultEnv
	}
	for _, v := range c.Sessions {
		if v.Labels.Env != env {
			continue
		}
		if v.Store == store {
			return v, nil
		}
	}
	return nil, errors.New("unknown SessionStore")
}

func (c *ConfigResolver) GetSlsConfig(cluster string) (*SlsOpt, error) {
	for _, sink := range c.EventSinks {
		if sink.Cluster != cluster {
			continue
		}
		return sink.SlsOpt, nil
	}
	return nil, errors.New("unknown config")
}

func (c *ConfigResolver) GetEventSinksFilters(cluster string) []string {
	var filters []string
	for _, sink := range c.EventSinks {
		if sink.Cluster != cluster {
			continue
		}
		return sink.Filters
	}
	return filters
}

func (c *ConfigResolver) GetEventSinksNotFilters(cluster string) []string {
	var filters []string
	for _, sink := range c.EventSinks {
		if sink.Cluster != cluster {
			continue
		}
		return sink.NotFilters
	}
	return filters
}
