package entity

type HostKey struct {
	Host        string
	KeyType     string
	Fingerprint string
	Key         string
}

func (h HostKey) SameKey(other HostKey) bool {
	return h.KeyType == other.KeyType && h.Key == other.Key
}
