package main

import (
	"flag"

	"github.com/BurntSushi/toml"
)

type Config struct {
	PrivateKey string `toml:"private_key"`
	Rpc        string `toml:"rpc"`
	Tick       string `toml:"tick"`
	Amt        int    `toml:"amt"`
	Prefix     string `toml:"prefix"`
	Count      int    `toml:"count"`
	GasTip     int    `toml:"gas_tip"`
	GasMax     int    `toml:"gas_max"`
	RealSend   bool   `toml:"realsend"`
}

var confPath = flag.String("c", "config.txt", "config file path")

var config *Config

func init() {
	flag.Parse()
	config = new(Config)
	_, err := toml.DecodeFile(*confPath, config)
	if err != nil {
		panic(err)
	}
}
