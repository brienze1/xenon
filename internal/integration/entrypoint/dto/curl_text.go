package dto

type CurlText struct {
	Text string `json:"text" validate:"required,not_blank"`
}

func (c CurlText) LogValue() any {
	return map[string]any{"length": len(c.Text)}
}
