package main

import "github.com/silentred/beim/comet"

func main() {
	s := comet.NewServer(nil)
	s.ListenAndServe(":1883")
}
