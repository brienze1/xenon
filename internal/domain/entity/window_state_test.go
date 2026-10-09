package entity

import "testing"

func TestScreenContainsItsAreaOnly(t *testing.T) {
	screen := Screen{X: 100, Y: 50, Width: 200, Height: 100}
	cases := map[[2]int]bool{
		{100, 50}: true, {299, 149}: true, {99, 50}: false, {100, 49}: false, {300, 50}: false, {100, 150}: false,
	}
	for point, want := range cases {
		if got := screen.Contains(point[0], point[1]); got != want {
			t.Errorf("Contains(%v) = %v, want %v", point, got, want)
		}
	}
}

func TestWindowStateFitsTheScreenItsCentreIsOn(t *testing.T) {
	left := Screen{X: -1000, Y: 0, Width: 1000, Height: 800}
	primary := Screen{X: 0, Y: 0, Width: 1600, Height: 1000, Primary: true}
	onEdge := WindowState{X: -500, Y: 0, Width: 1000, Height: 700, Placed: true}
	if got := onEdge.FittedTo([]Screen{left, primary}); got.X != 0 {
		t.Errorf("a window centred on x=0 belongs to the primary screen, got %+v", got)
	}
	inside := WindowState{X: -501, Y: 0, Width: 1000, Height: 700, Placed: true}
	if got := inside.FittedTo([]Screen{left, primary}); got.X != -1000 || got.Width != 1000 {
		t.Errorf("a window centred left of x=0 stays on the left screen, got %+v", got)
	}
	below := WindowState{X: 0, Y: 650, Width: 1000, Height: 700, Placed: true}
	tall := Screen{X: 0, Y: 1000, Width: 1600, Height: 1000}
	if got := below.FittedTo([]Screen{primary, tall}); got.Y != 1000 {
		t.Errorf("a window centred on y=1000 belongs to the lower screen, got %+v", got)
	}
}

func TestWindowStateCentresAnUnplacedWindow(t *testing.T) {
	got := NewWindowState().FittedTo([]Screen{{X: 10, Y: 20, Width: 1290, Height: 811, Primary: true}})
	if got != (WindowState{X: 15, Y: 25, Width: 1280, Height: 800}) {
		t.Errorf("FittedTo() = %+v", got)
	}
}

func TestWindowStateKeepsNormalBoundsWhileMaximised(t *testing.T) {
	normal := WindowState{X: 1, Y: 2, Width: 1000, Height: 700, Placed: true}
	maximised := normal.Updated(WindowState{Width: 1920, Height: 1080, Maximized: true})
	if maximised != (WindowState{X: 1, Y: 2, Width: 1000, Height: 700, Maximized: true, Placed: true}) {
		t.Errorf("Updated(maximised) = %+v", maximised)
	}
	restored := maximised.Updated(WindowState{X: 5, Y: 6, Width: 1100, Height: 750})
	if restored != (WindowState{X: 5, Y: 6, Width: 1100, Height: 750, Placed: true}) {
		t.Errorf("Updated(normal) = %+v", restored)
	}
}
