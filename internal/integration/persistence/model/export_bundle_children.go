package model

import "go.yaml.in/yaml/v3"

type ExportBundleChildren struct {
	Children []yaml.Node `yaml:"children"`
}
