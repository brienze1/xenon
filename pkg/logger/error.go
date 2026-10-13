package logger

type codedError interface {
	Code() string
	Description() string
	StatusCode() int
}

type errorModel struct {
	Message     string `json:"message"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
}

func newErrorModel(err error) *errorModel {
	if err == nil {
		return nil
	}
	model := &errorModel{Message: err.Error()}
	if coded, ok := err.(codedError); ok {
		model.Code = coded.Code()
		model.Description = coded.Description()
		model.StatusCode = coded.StatusCode()
	}
	return model
}
