package cmd

import (
	"reflect"
	"strings"

	"github.com/deltacat/dbstress/cmd/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func initConfig() {
	project := version.Project

	viper.SetConfigName(project)
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/" + project)
	viper.AddConfigPath("/etc/" + project)
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			logrus.Warning("No configuration file found, using defaults.")
		default:
			logrus.WithError(err).Fatal("read configuration file error")
		}
	} else {
		logrus.Info("configuration file loaded: ", viper.ConfigFileUsed())
	}

	viperBindEnvs(config.Cfg)

	if err := viper.Unmarshal(&config.Cfg); err != nil {
		logrus.WithError(err).Fatal("unmarshal config error")
	}
}

func viperBindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			tv = strings.ToLower(t.Name)
		}
		if tv == "-" {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			viperBindEnvs(v.Interface(), append(parts, tv)...)
		default:
			key := strings.Join(append(parts, tv), ".")
			if err := viper.BindEnv(key); err != nil {
				logrus.Error(err)
			}
		}
	}
}

func setDefaultConfig() {
	viper.SetDefault("connection.influxdb", "http://127.0.0.1:8086")
}
