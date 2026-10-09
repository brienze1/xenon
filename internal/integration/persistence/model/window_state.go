package model

import "github.com/brienze1/xenon/internal/domain/entity"

type WindowState struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximized bool `json:"maximized"`
}

func NewWindowState(state entity.WindowState) WindowState {
	return WindowState{X: state.X, Y: state.Y, Width: state.Width, Height: state.Height, Maximized: state.Maximized}
}

func (w *WindowState) ToEntity() entity.WindowState {
	return entity.WindowState{X: w.X, Y: w.Y, Width: w.Width, Height: w.Height, Maximized: w.Maximized, Placed: true}
}
