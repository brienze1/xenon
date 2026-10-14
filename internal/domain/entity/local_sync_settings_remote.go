package entity

// Remote is the repository URL safe to show: without embedded credentials.
func (l LocalSyncSettings) Remote() string {
	return RemoteWithoutCredentials(l.Repository)
}
