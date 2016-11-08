package comet

var serviceMap = make(map[string]*service)

func getService(name string) *service {
	if s, ok := serviceMap[name]; ok {
		return s
	}
	return nil
}

// TODO: thread safe, rw lock
func saveService(name string, service *service) {
	serviceMap[name] = service
}
