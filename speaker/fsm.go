package speaker

type FSM interface {
	Send(int)
}

type DefaultFSM struct {
	fsm *kbgp.FSM
}

func NewDefaultFSM() *DefaultFSM {
	return &DefaultFSM{}
}

func (f *DefaultFSM) Send(event int) {
}
