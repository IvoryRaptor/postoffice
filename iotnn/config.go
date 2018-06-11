package iotnn


type ZookeeperConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Config struct {
	OAuth     string          `yaml:"oAuth"`
	Zookeeper ZookeeperConfig `yaml:"zookeeper"`
}
