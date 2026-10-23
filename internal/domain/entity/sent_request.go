package entity

type SentRequest struct {
	Method  string
	URL     string
	Body    string
	Headers []KeyValue
}

func NewSentRequest(request EffectiveRequest) SentRequest {
	return SentRequest{Method: request.Method, URL: request.URL, Headers: request.SendHeaders(), Body: request.Body.PreviewText()}
}

func (s SentRequest) WithSecretsMasked(mask func(text string) (string, bool)) SentRequest {
	masker := secretMasker{mask: mask}
	s.URL = masker.text(s.URL)
	s.Body = masker.text(s.Body)
	s.Headers = mapKeyValues(s.Headers, masker.text)
	return s
}
