package dto

type Example struct {
	Extra    map[string]any  `yaml:",inline"`
	Order    *float64        `yaml:"order"`
	ID       any             `yaml:"id"`
	Kind     string          `yaml:"$kind"`
	Name     string          `yaml:"name"`
	Request  ExampleRequest  `yaml:"request"`
	Response ExampleResponse `yaml:"response"`
}

func (e *Example) toPostmanV21() (map[string]any, error) {
	response := copied(e.Extra)
	response["name"] = e.Name
	setIfPresent(response, "id", e.ID)
	request, err := e.Request.toPostmanV21()
	if err != nil {
		return nil, err
	}
	response["originalRequest"] = request
	for key, value := range e.Response.Extra {
		response[key] = value
	}
	response["code"] = e.Response.StatusCode
	response["status"] = e.Response.StatusText
	response["header"] = e.Response.Headers.toPostmanV21()
	if e.Response.Body != nil {
		response["body"] = e.Response.Body.contentText()
		setIfPresent(response, "_postman_previewlanguage", e.Response.Body.Type)
	}
	return response, nil
}
