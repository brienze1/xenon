package entity

const (
	MinWindowWidth      = 960
	MinWindowHeight     = 600
	defaultWindowWidth  = 1280
	defaultWindowHeight = 800
)

type WindowState struct {
	X         int
	Y         int
	Width     int
	Height    int
	Maximized bool
	Placed    bool
}

func NewWindowState() WindowState {
	return WindowState{Width: defaultWindowWidth, Height: defaultWindowHeight}
}

func (w WindowState) Updated(next WindowState) WindowState {
	if next.Maximized {
		w.Maximized = true
		return w
	}
	next.Placed = true
	return next
}

func (w WindowState) FittedTo(screens []Screen) WindowState {
	screen := w.screenOf(screens)
	w.Width = fitSize(w.Width, MinWindowWidth, screen.Width)
	w.Height = fitSize(w.Height, MinWindowHeight, screen.Height)
	if !w.Placed {
		w.X = screen.X + (screen.Width-w.Width)/2
		w.Y = screen.Y + (screen.Height-w.Height)/2
	}
	w.X = min(max(w.X, screen.X), screen.X+screen.Width-w.Width)
	w.Y = min(max(w.Y, screen.Y), screen.Y+screen.Height-w.Height)
	return w
}

func (w WindowState) screenOf(screens []Screen) Screen {
	if w.Placed {
		for _, screen := range screens {
			if screen.Contains(w.X+w.Width/2, w.Y+w.Height/2) {
				return screen
			}
		}
	}
	for _, screen := range screens {
		if screen.Primary {
			return screen
		}
	}
	return screens[0]
}

func fitSize(size, minimum, available int) int {
	return min(max(size, min(minimum, available)), available)
}
