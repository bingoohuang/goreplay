package main

import (
	"os"
)

// DummyOutput used for debugging, prints all incoming requests
type DummyOutput struct {
}

// NewDummyOutput constructor for DummyOutput
func NewDummyOutput() (di *DummyOutput) {
	di = new(DummyOutput)

	return
}

// PluginWrite writes message to this plugin
func (i *DummyOutput) PluginWrite(msg *Message) (int, error) {
	var n1, n2, n3, n4 int
	var err error
	n1, err = os.Stdout.WriteString(msg.Key + " ")
	n2, err = os.Stdout.Write(msg.Meta)
	n3, err = os.Stdout.Write(msg.Data)
	n4, err = os.Stdout.Write(payloadSeparatorAsBytes)
	return n1 + n2 + n3 + n4, err
}

func (i *DummyOutput) String() string {
	return "Dummy Output"
}
