package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Screen struct {
	X       int  `json:"x"`
	Y       int  `json:"y"`
	Width   int  `json:"width" validate:"gt=0"`
	Height  int  `json:"height" validate:"gt=0"`
	Primary bool `json:"primary"`
}

func NewScreen(x, y, width, height int, primary bool) Screen {
	return Screen{X: x, Y: y, Width: width, Height: height, Primary: primary}
}

func (s *Screen) ToEntity() entity.Screen {
	return entity.Screen{X: s.X, Y: s.Y, Width: s.Width, Height: s.Height, Primary: s.Primary}
}
