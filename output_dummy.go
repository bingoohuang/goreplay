package main

import (
	"os"
	"strconv"
	"strings"
	"time"
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
	n2, err = os.Stdout.WriteString(prettyMeta(msg.Meta))
	n3, err = os.Stdout.Write(msg.Data)
	n4, err = os.Stdout.Write(payloadSeparatorAsBytes)
	return n1 + n2 + n3 + n4, err
}

func prettyMeta(meta []byte) string {
	//func payloadHeader(payloadType byte, uuid []byte, timing int64, latency int64) (header []byte) {
	//	//Example:
	//	//  3 f45590522cd1838b4a0d5c5aab80b77929dea3b3 13923489726487326 1231\n
	//	return []byte(fmt.Sprintf("%c %s %d %d\n", payloadType, uuid, timing, latency))
	//}
	fields := strings.Fields(string(meta))
	// payloadType 1 - request, 2 - response, 3 - replayed response
	switch payloadType := fields[0]; payloadType {
	case "1":
		fields[0] = "request"
	case "2":
		fields[0] = "response"
	case "3":
		fields[0] = "replayed response"
	}
	nano, _ := strconv.ParseInt(fields[2], 10, 64)
	fields[2] += time.Unix(0, nano).Format(`(2006-01-02 15:04:05.000000)`)

	return strings.Join(fields, " ") + "\n"
}

func (i *DummyOutput) String() string {
	return "Dummy Output"
}
