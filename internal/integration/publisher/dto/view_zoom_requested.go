package dto

const ViewZoomRequestedEvent = "view:zoom-requested"

// ViewZoomRequested asks the UI to zoom: "in", "out" or "reset".
type ViewZoomRequested struct {
	Direction string `json:"direction"`
}

func NewViewZoomRequested(direction string) ViewZoomRequested {
	return ViewZoomRequested{Direction: direction}
}
