package config

import (
	"os"
	"strings"

	"github.com/josexy/gsocks5/socks/auth"
	"github.com/josexy/gsocks5/socks/constant"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	ListenAddr  string
	SocksMethod []constant.Socks5Method
	Auth        []auth.Socks5Auth
}

type yamlConfig struct {
	ListenAddr  string   `yaml:"listen_addr"`
	SocksMethod []string `yaml:"socks_method"`
	Auth        []string `yaml:"auth"`
}

var Cfg = new(AppConfig)

func ParseConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	cfg := new(yamlConfig)
	if err = yaml.Unmarshal(data, cfg); err != nil {
		panic(err)
	}
	Cfg.ListenAddr = cfg.ListenAddr
	for _, method := range cfg.SocksMethod {
		switch method {
		case "username":
			Cfg.SocksMethod = append(Cfg.SocksMethod, constant.MethodUsernamePassword)
		case "none":
			Cfg.SocksMethod = append(Cfg.SocksMethod, constant.MethodNoAuthRequired)
		}
	}
	for _, x := range cfg.Auth {
		parts := strings.Split(x, ":")
		Cfg.Auth = append(Cfg.Auth, auth.NewSocksAuth(parts[0], parts[1]))
	}
}
