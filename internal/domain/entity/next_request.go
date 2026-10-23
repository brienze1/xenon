package entity

// NextRequest is a script's pm.execution.setNextRequest call: a target request id or name, or Stop for null.
type NextRequest struct {
	Target string
	Stop   bool
}

func NewNextRequest(target string) NextRequest {
	return NextRequest{Target: target}
}

func NewStopNextRequest() NextRequest {
	return NextRequest{Stop: true}
}
