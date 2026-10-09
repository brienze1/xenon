package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type WindowState struct {
	X          int  `json:"x"`
	Y          int  `json:"y"`
	Width      int  `json:"width" validate:"gt=0"`
	Height     int  `json:"height" validate:"gt=0"`
	Maximized  bool `json:"maximized"`
	Minimised  bool `json:"minimised"`
	Fullscreen bool `json:"fullscreen"`
}

func NewWindowState(state entity.WindowState) WindowState {
	return WindowState{X: state.X, Y: state.Y, Width: state.Width, Height: state.Height, Maximized: state.Maximized}
}

func NewWindowStateFromBounds(x, y, width, height int, maximized, minimised, fullscreen bool) WindowState {
	return WindowState{X: x, Y: y, Width: width, Height: height, Maximized: maximized, Minimised: minimised, Fullscreen: fullscreen}
}

func (w *WindowState) ToEntity() entity.WindowState {
	return entity.WindowState{X: w.X, Y: w.Y, Width: w.Width, Height: w.Height, Maximized: w.Maximized, Minimised: w.Minimised, Fullscreen: w.Fullscreen}
}
