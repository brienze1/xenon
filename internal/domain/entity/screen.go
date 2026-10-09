package entity

type Screen struct {
	X       int
	Y       int
	Width   int
	Height  int
	Primary bool
}

func (s Screen) Contains(x, y int) bool {
	return x >= s.X && x < s.X+s.Width && y >= s.Y && y < s.Y+s.Height
}
