package main

import (
	"networking"
)

func main() {
	simpleNode := networking.NewSimpleNode("192.168.3.6", 8333, false)
	simpleNode.Run()
}
