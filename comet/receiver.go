package comet

type ReceiverProvider interface {
	Receive() ([]byte, error)
}
