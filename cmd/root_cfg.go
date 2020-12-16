package cmd

import (
	"reflect"
	"strings"

	"github.com/deltacat/dbstress/config"
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
	viper.SetDefault("connection.influxdb.url", "http://127.0.0.1:8086")
	viper.SetDefault("connection.influxdb.user", "")
	viper.SetDefault("connection.influxdb.pass", "")
	viper.SetDefault("connection.influxdb.db", "stress")
	viper.SetDefault("connection.influxdb.rp", "")
	viper.SetDefault("connection.influxdb.precision", "n")
	viper.SetDefault("connection.influxdb.consistency", "one")
	viper.SetDefault("connection.influxdb.tls-skip-verify", false)
	viper.SetDefault("connection.influxdb.gzip", -1)

	viper.SetDefault("points.series-key", "ctr,some=tag")
	viper.SetDefault("points.fields-str", "n=0i")
}
