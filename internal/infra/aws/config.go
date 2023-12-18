package aws

type Configs []Config

type Config struct {
	Region string `yaml:"region"`
}
