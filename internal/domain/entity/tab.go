package entity

type Tab struct {
	Draft        *Request
	TabID        string
	RequestID    string
	BaseRevision string
	Pinned       bool
}

func (t Tab) LogValue() any {
	value := map[string]any{"tabId": t.TabID, "requestId": t.RequestID, "baseRevision": t.BaseRevision, "pinned": t.Pinned}
	if t.Draft != nil {
		value["draft"] = t.Draft.LogValue()
	}
	return value
}
