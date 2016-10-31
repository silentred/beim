package main

import "beim/comet"

func main() {
	s := comet.NewServer(nil)
	s.ListenAndServe(":1883")
}
