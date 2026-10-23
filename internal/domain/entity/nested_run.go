package entity

const (
	maxNestedRunDepth = 10
	nestedRunNotFound = "runRequest: no saved request with id "
	nestedRunTooDeep  = "runRequest: requests run from scripts are nested too deep"
)

// NestedRun is a pm.execution.runRequest call made at a nesting depth.
type NestedRun struct {
	RequestID string
	Depth     int
}

func NewNestedRun(requestId string, depth int) NestedRun {
	return NestedRun{RequestID: requestId, Depth: depth}
}

// Rejection refuses the call before the saved request is read: when it is nested too deep or the id is not a saved request.
func (n NestedRun) Rejection() (NestedRequest, bool) {
	if n.Depth >= maxNestedRunDepth {
		return NewMissingRunNestedRequest(n.RequestID, nestedRunTooDeep), true
	}
	if !IsSavedRequestID(n.RequestID) {
		return n.NotFound(), true
	}
	return NestedRequest{}, false
}

func (n NestedRun) NotFound() NestedRequest {
	return NewMissingRunNestedRequest(n.RequestID, nestedRunNotFound+n.RequestID)
}
