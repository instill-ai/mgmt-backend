package worker

const TaskQueue = "mgmt-backend"

type Worker interface {
}

type worker struct {
}

func NewWorker() Worker {
	return &worker{}
}
