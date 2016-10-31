package comet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/surgemq/message"
)

var (
	ErrInvalidConnectionType = errors.New("invalid connection type")
)

func getConnectMessage(conn io.Closer) (*message.ConnectMessage, error) {
	buf, err := getMessageBuffer(conn)
	if err != nil {
		//glog.Debugf("Receive error: %v", err)
		return nil, err
	}

	msg := message.NewConnectMessage()

	_, err = msg.Decode(buf)
	//glog.Debugf("Received: %s", msg)
	return msg, err
}

func getConnackMessage(conn io.Closer) (*message.ConnackMessage, error) {
	buf, err := getMessageBuffer(conn)
	if err != nil {
		//glog.Debugf("Receive error: %v", err)
		return nil, err
	}

	msg := message.NewConnackMessage()

	_, err = msg.Decode(buf)
	//glog.Debugf("Received: %s", msg)
	return msg, err
}

func writeMessage(conn io.Closer, msg message.Message) error {
	buf := make([]byte, msg.Len())
	_, err := msg.Encode(buf)
	if err != nil {
		//glog.Debugf("Write error: %v", err)
		return err
	}
	//glog.Debugf("Writing: %s", msg)

	return writeMessageBuffer(conn, buf)
}

func getMessageBuffer(c io.Closer) ([]byte, error) {
	if c == nil {
		return nil, ErrInvalidConnectionType
	}

	conn, ok := c.(io.Reader)
	if !ok {
		return nil, ErrInvalidConnectionType
	}

	var (
		// the message buffer
		buf []byte

		// tmp buffer to read a single byte
		b []byte = make([]byte, 1)

		// total bytes read
		l int = 0
	)

	// Let's read enough bytes to get the message header (msg type, remaining length)
	for {
		// If we have read 5 bytes and still not done, then there's a problem.
		if l > 5 {
			return nil, fmt.Errorf("connect/getMessage: 4th byte of remaining length has continuation bit set")
		}

		n, err := conn.Read(b[0:])
		if err != nil {
			//glog.Debugf("Read error: %v", err)
			return nil, err
		}

		// Technically i don't think we will ever get here
		if n == 0 {
			continue
		}

		buf = append(buf, b...)
		l += n

		// Check the remlen byte (1+) to see if the continuation bit is set. If so,
		// increment cnt and continue reading. Otherwise break.
		if l > 1 && b[0] < 0x80 {
			break
		}
	}

	// Get the remaining length of the message
	remlen, _ := binary.Uvarint(buf[1:])
	buf = append(buf, make([]byte, remlen)...)

	for l < len(buf) {
		n, err := conn.Read(buf[l:])
		if err != nil {
			return nil, err
		}
		l += n
	}

	return buf, nil
}

func writeMessageBuffer(c io.Closer, b []byte) error {
	if c == nil {
		return ErrInvalidConnectionType
	}

	conn, ok := c.(io.Writer)
	if !ok {
		return ErrInvalidConnectionType
	}

	_, err := conn.Write(b)
	return err
}

// Copied from http://golang.org/src/pkg/net/timeout_test.go
func isTimeout(err error) bool {
	e, ok := err.(net.Error)
	return ok && e.Timeout()
}
