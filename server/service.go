package server

// Service represents a startable & stoppable service.
type Service interface {
	ListenAndServe() error
	GracefulStop()
}
