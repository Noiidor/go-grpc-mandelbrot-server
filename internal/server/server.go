package server

type Server interface {
	Serve() error
	Stop()
	GetServer() interface{}
}
