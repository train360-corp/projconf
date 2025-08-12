package platform

type Service interface {
	Start() error
	Stop() error
}
