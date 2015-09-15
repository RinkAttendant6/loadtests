package scheduler

type DB interface {
	SetExecutorUp(dropletID int, port int) error
}
