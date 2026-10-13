package adapter

type Keychain interface {
	Get(service, account string) (string, bool, error)
	Set(service, account, secret string) error
	Delete(service, account string) error
}
