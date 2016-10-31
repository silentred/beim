package comet

type Authenticator interface {
	Authenticate(username, password string) bool
}

type mockAuth struct {
}

func (m *mockAuth) Authenticate(username, password string) bool {
	return true
}
