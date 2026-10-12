package dto

type Server struct {
	Url         string `yaml:"url"`
	Description string `yaml:"description,omitempty"`
}

func NewServers(url string) []Server {
	if url == "" {
		return nil
	}
	return []Server{{Url: url}}
}

func firstServer(servers []Server) string {
	if len(servers) == 0 {
		return ""
	}
	return servers[0].Url
}

func (s Server) MarshalYAML() (any, error) {
	return mappingOfFields(always("url", s.Url), text("description", s.Description))
}
