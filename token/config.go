package token

import "github.com/spf13/pflag"

type Config struct {
	Name   string `json:"name" yaml:"name"`
	Secret string `json:"secret" yaml:"secret"`
}

func (c *Config) FlagSet() *pflag.FlagSet {
	return FlagSet()
}

func FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("token", pflag.ContinueOnError)
	fs.String("token.name", "token", "Token Name")
	fs.String("token.secret", "", "Token Secret")
	return fs
}
