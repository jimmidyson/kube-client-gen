package generator

import (
	"github.com/inconshreveable/log15"

	"github.com/jimmidyson/kube-client-gen/pkg/loader"
)

type Config struct {
	Logger          log15.Logger
	Force           bool
	OutputDirectory string
}

type Generator interface {
	Generate([]loader.Package) error
}
