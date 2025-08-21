package config

import (
	"fmt"
	"log"
	"os"

	"reflect"

	"gopkg.in/yaml.v3"
)

var CommandPrefix = "+"
var BotMode = "+B"
var Server = "irc.zoite.net:6697"
var Channel = "#antisocial"
var MaxMessagePool = 20
var DeletionDays = 1
var MessageQuota = 400
var PeopleQuota = 5
var Bert = true
var GPU = true

type BotStruct struct {
	Prefix  string `yaml:"prefix"`
	Mode    string `yaml:"mode"`
	Server  string `yaml:"server"`
	Channel string `yaml:"channel"`
}

type StorageStruct struct {
	MessagePoolSize int `yaml:"message_pool_size"`
	MessageQuota    int `yaml:"message_quota"`
	PeopleQuota     int `yaml:"people_quota"`
}

type SchedulerStruct struct {
	DeletionDays int `yaml:"deletion_days"`
}

type ModelStruct struct {
	Bert bool `yaml:"bert"`
	GPU  bool `yaml:"gpu"`
}

type ConfigStruct struct {
	Bot       BotStruct       `yaml:"bot"`
	Storage   StorageStruct   `yaml:"storage"`
	Scheduler SchedulerStruct `yaml:"scheduler"`
	Model     ModelStruct     `yaml:"model"`
}

func List(v interface{}) {
	value := reflect.ValueOf(v)
	typeR := reflect.TypeOf(v)

	if value.Kind() == reflect.Pointer {
		value = value.Elem()
		typeR = typeR.Elem()
	}

	for i := 0; i < value.NumField(); i++ {
		field := typeR.Field(i)
		val := value.Field(i)

		if val.Kind() == reflect.Struct {
			List(val.Interface())
		} else {
			fmt.Printf("%s: %v\n", field.Name, val.Interface())
		}
	}
}

func ReadConfig(path string, verbose bool) error {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to open config file: %s\n", err)
		return err
	}

	var cfg ConfigStruct
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.Printf("Failed to unmarshal YAML: %s\n", err)
		return err
	}

	if cfg.Bot.Prefix != "" {
		CommandPrefix = cfg.Bot.Prefix
	}
	if cfg.Bot.Mode != "" {
		BotMode = cfg.Bot.Mode
	}
	if cfg.Bot.Server != "" {
		Server = cfg.Bot.Server
	}
	if cfg.Bot.Channel != "" {
		Channel = cfg.Bot.Channel
	}

	if cfg.Storage.MessagePoolSize > 0 {
		MaxMessagePool = cfg.Storage.MessagePoolSize
	}
	if cfg.Storage.MessageQuota > 0 {
		MessageQuota = cfg.Storage.MessageQuota
	}
	if cfg.Storage.PeopleQuota > 0 {
		PeopleQuota = cfg.Storage.PeopleQuota
	}

	if cfg.Scheduler.DeletionDays > 0 {
		DeletionDays = cfg.Scheduler.DeletionDays
	}

	Bert = cfg.Model.Bert
	GPU = cfg.Model.GPU

	if verbose {
		List(cfg)
	}

	return nil
}
