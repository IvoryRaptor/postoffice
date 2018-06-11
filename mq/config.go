package mq


type Config struct {
	Type string	`yaml:"type"`
	Host string	`yaml:"host"`
	Port int	`yaml:"port"`
	Other map[string]interface{} `yaml:"ohter"`
}
