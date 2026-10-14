package dto

type ExampleRequest struct {
	Extra       map[string]any `yaml:",inline"`
	Body        *Body          `yaml:"body"`
	URL         string         `yaml:"url"`
	Method      string         `yaml:"method"`
	Headers     KeyValues      `yaml:"headers"`
	QueryParams KeyValues      `yaml:"queryParams"`
}

func (e *ExampleRequest) toPostmanV21() (map[string]any, error) {
	request := copied(e.Extra)
	request["method"] = e.Method
	request["header"] = e.Headers.toPostmanV21()
	request["url"] = NewURL(e.URL, e.QueryParams, nil).toPostmanV21()
	body, err := e.Body.toPostmanV21()
	if err != nil {
		return nil, err
	}
	setIfPresent(request, "body", body)
	return request, nil
}
