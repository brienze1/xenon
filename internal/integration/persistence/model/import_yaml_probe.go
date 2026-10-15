package model

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportYamlProbe struct {
	ImportProbe `yaml:",inline"`
}

func (i *ImportYamlProbe) ToEntity(name string) entity.ImportProbe {
	probe := i.postmanProbe(name)
	probe.IsYaml = true
	return probe
}
