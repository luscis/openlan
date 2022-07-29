package libol

type WaitOne struct {
	done chan bool
}

func NewWaitOne(n int) *WaitOne {
	return &WaitOne{
		done: make(chan bool, n),
	}
}

func (w *WaitOne) Done() {
	w.done <- true
}

func (w *WaitOne) Wait() {
	<-w.done
}
