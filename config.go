package main

import (
    "io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

var config Config

type Config struct {
	OldDB string
	NewDB string
    Database DatabaseConfig
}

type DatabaseConfig struct {
	Host string
	Username string
	Password string
}

func CreateNewConfig() {
	newConfig := `# MyHomeworkSpace migration configuration
OldDB = "myhomeworkspace_old"
NewDB = "myhomeworkspace_new"

[Database]
Host = "localhost:3306"
Username = "myhomeworkspace"
Password = "myhomeworkspace"`
	err := ioutil.WriteFile("config.toml", []byte(newConfig), 0644)
	if err != nil {
		panic(err)
	}
}

func InitConfig() {
	if _, err := os.Stat("config.toml"); err != nil {
		CreateNewConfig() // create new config to be parsed
		panic("No config found, please edit the generated one")
	}
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		panic(err)
	}
}
