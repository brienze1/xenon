package entity

type CookieJarState struct {
	Cookies       []Cookie
	Revision      uint64
	SavedRevision uint64
	Loaded        bool
}

func (c CookieJarState) HasUnsavedChanges() bool {
	return c.Revision != c.SavedRevision
}

func (c CookieJarState) ChangedSince(revision uint64) bool {
	return c.Revision != revision
}
