package config

type Config interface {
	GetFromGRPC(key string) (string, error)
}
