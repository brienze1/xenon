package dto

type ExampleResponse struct {
	Extra      map[string]any `yaml:",inline"`
	Body       *Body          `yaml:"body"`
	StatusText string         `yaml:"statusText"`
	Headers    KeyValues      `yaml:"headers"`
	StatusCode int            `yaml:"statusCode"`
}
