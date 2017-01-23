package lib

type Map map[string]interface{}

func (s *Map) Set(key string, val interface{}) {
	(*s)[key] = val
}

func (s *Map) Get(key string) interface{} {
	if val, ok := (*s)[key]; ok {
		return val
	}
	return nil
}

func (s *Map) Delete(key string) {
	delete(*s, key)
}
