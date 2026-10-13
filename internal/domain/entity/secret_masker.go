package entity

type secretMasker struct {
	mask   func(text string) (string, bool)
	masked bool
}

func (s *secretMasker) text(text string) string {
	masked, replaced := s.mask(text)
	s.masked = s.masked || replaced
	return masked
}

func (s *secretMasker) texts(texts []string) []string {
	masked := make([]string, 0, len(texts))
	for _, text := range texts {
		masked = append(masked, s.text(text))
	}
	return masked
}

func (s *secretMasker) values(values []any) []any {
	if values == nil {
		return nil
	}
	masked := make([]any, 0, len(values))
	for _, value := range values {
		masked = append(masked, s.value(value))
	}
	return masked
}

func (s *secretMasker) value(value any) any {
	switch typed := value.(type) {
	case string:
		return s.text(typed)
	case []any:
		return s.values(typed)
	case map[string]any:
		masked := make(map[string]any, len(typed))
		for key, item := range typed {
			masked[s.text(key)] = s.value(item)
		}
		return masked
	default:
		return value
	}
}

func (s *secretMasker) request(request EffectiveRequest) EffectiveRequest {
	request.URL = s.text(request.URL)
	request.Headers = mapKeyValues(request.Headers, s.text)
	autoHeaders := make([]AutoHeader, 0, len(request.AutoHeaders))
	for _, header := range request.AutoHeaders {
		header.Value = s.text(header.Value)
		autoHeaders = append(autoHeaders, header)
	}
	request.AutoHeaders = autoHeaders
	request.Body = request.Body.MapText(s.text)
	request.Auth = request.Auth.MapText(s.text)
	return request
}

func (s *secretMasker) response(response Response) Response {
	response.Headers = mapKeyValues(response.Headers, s.text)
	if !response.Binary && len(response.Body) > 0 {
		response.Body = []byte(s.text(string(response.Body)))
	}
	if response.Error != nil {
		sendError := *response.Error
		sendError.Message = s.text(sendError.Message)
		response.Error = &sendError
	}
	return response
}
