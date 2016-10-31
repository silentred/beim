package comet

import "testing"

func TestMain(t *testing.T) {
	s := NewServer(nil)
	s.ListenAndServe(":1883")
}
