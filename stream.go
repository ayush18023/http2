package main

type Stream struct {
	frames  chan *Frame
	quitStr chan struct{}
}
