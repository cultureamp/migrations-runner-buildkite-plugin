package config

import "github.com/kelseyhightower/envconfig"

type RunParameters struct {
	ParameterName string `split_words:"true" required:"true"`
	Script        string `split_words:"true" required:"true"`
}

func New() *RunParameters {
	rp := RunParameters{}
	envconfig.MustProcess("", &rp)

	return &rp
}
