package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// RestOutputConfig struct for holding http output configuration
type RestOutputConfig struct {
	FlushInterval time.Duration `json:"output-rest-flush-interval"`
}

// RestOutput output plugin
type RestOutput struct {
	config   *RestOutputConfig
	filename string
	file     *os.File
	writer   io.Writer

	sync.RWMutex

	stop    chan struct{}
	flusher func() error
}

// NewRestOutput constructor for HTTPOutput
// Initialize workers
func NewRestOutput(filename string) *RestOutput {
	o := new(RestOutput)
	o.filename = filename
	o.config = &RestOutputConfig{
		FlushInterval: 100 * time.Millisecond,
	}
	o.stop = make(chan struct{})

	go func() {
		ticker := time.NewTicker(o.config.FlushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				o.flush()
			case <-o.stop:
				o.flush()
				o.file.Close()
				o.stop <- struct{}{}
				return
			}
		}
	}()

	return o
}

// PluginWrite writes message to this plugin
func (o *RestOutput) PluginWrite(msg *Message) (n int, err error) {
	if !isRequestPayload(msg.Meta) {
		return len(msg.Data), nil
	}

	var req *http.Request
	req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Data)))
	if err != nil {
		return n, err
	}
	// we don't send CONNECT or OPTIONS request
	if req.Method == http.MethodConnect {
		return n, nil
	}

	// avoid replay
	if req.Header.Get("X-Goreplay-Output") == "1" {
		return n, nil
	}

	return o.writeMessage(msg, req)
}

func (o *RestOutput) writeMessage(msg *Message, req *http.Request) (n int, err error) {
	o.RLock()
	defer o.RUnlock()

	o.createFile()

	// https://www.jetbrains.com/help/idea/exploring-http-syntax.html#compose-several-requests-in-a-file
	var n1 int
	n1, err = o.writer.Write([]byte("# " + msg.Key + " " + prettyMeta(msg.Meta)))
	n += n1

	req.TransferEncoding = []string{"identity"}
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return n, err
	}

	identity := []byte("\r\nTransfer-Encoding: identity")
	n1, err = o.writer.Write(bytes.Replace(dump, identity, nil, -1))
	n += n1

	n1, err = o.writer.Write([]byte("\n\n###\n\n"))
	n += n1

	return n, nil
}

func (o *RestOutput) String() string {
	return "Rest output: " + o.filename
}

// Close closes the data channel so that data
func (o *RestOutput) Close() error {
	o.stop <- struct{}{}
	<-o.stop

	return nil
}

func (o *RestOutput) createFile() {
	if o.file != nil {
		return
	}

	var err error

	o.file, err = os.OpenFile(o.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		log.Fatal(o, "Cannot open file %q. Error: %s", o.filename, err)
	}
	o.file.Sync()

	if strings.HasSuffix(o.filename, ".gz") {
		w := gzip.NewWriter(o.file)
		o.writer = w
		o.flusher = w.Flush
	} else {
		w := bufio.NewWriter(o.file)
		o.writer = w
		o.flusher = w.Flush
	}
}

func (o *RestOutput) flush() {
	// Don't exit on panic
	defer func() {
		if r := recover(); r != nil {
			Debug(0, "[OUTPUT-FILE] PANIC while file flush: ", r, o, string(debug.Stack()))
		}
	}()

	o.Lock()
	defer o.Unlock()

	if o.flusher != nil {
		_ = o.flusher()
	}
}
