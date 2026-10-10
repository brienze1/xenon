package dto

type YamlFolder struct {
	Docs    string              `yaml:"docs"`
	Info    YamlInfo            `yaml:"info"`
	Request YamlRequestSettings `yaml:"request"`
}

func (y *YamlFolder) ToFolder(file, fallbackName string) Folder {
	folder := Folder{
		Name:    y.Info.Name,
		Seq:     y.Info.Seq,
		File:    file,
		Docs:    trimFinalNewline(y.Docs),
		Headers: yamlPairs(y.Request.Headers),
		Auth:    yamlAuth(y.Request.Auth),
		Scripts: yamlScripts(y.Request.Scripts, y.Request.Variables, y.Request.Asserts),
	}
	if folder.Name == "" {
		folder.Name = fallbackName
	}
	return folder
}

type YamlCollection struct {
	Docs           string              `yaml:"docs"`
	OpenCollection string              `yaml:"opencollection"`
	Info           YamlInfo            `yaml:"info"`
	Request        YamlRequestSettings `yaml:"request"`
}

func (y *YamlCollection) ToCollection(file string) Collection {
	scripts := yamlScripts(y.Request.Scripts, y.Request.Variables, y.Request.Asserts)
	collection := Collection{
		Name:      y.Info.Name,
		File:      file,
		Docs:      trimFinalNewline(y.Docs),
		Headers:   yamlPairs(y.Request.Headers),
		Auth:      yamlAuth(y.Request.Auth),
		Variables: scripts.PreRequestVars,
		Scripts:   scripts,
		Config:    map[string]any{"opencollection": y.OpenCollection},
	}
	collection.Scripts.PreRequestVars = []Pair{}
	return collection
}

type YamlEnvironment struct {
	Name      string     `yaml:"name"`
	Color     string     `yaml:"color"`
	Variables []YamlPair `yaml:"variables"`
}

func (y *YamlEnvironment) ToEnvironment(fallbackName string) Environment {
	environment := Environment{Name: y.Name, Color: y.Color, Variables: []EnvironmentVariable{}}
	if environment.Name == "" {
		environment.Name = fallbackName
	}
	for _, variable := range y.Variables {
		environment.Variables = append(environment.Variables, EnvironmentVariable{Key: variable.Name, Value: variable.Value, Enabled: variable.IsEnabled(), Secret: variable.Secret})
	}
	return environment
}
