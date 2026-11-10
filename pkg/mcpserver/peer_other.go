//go:build !linux && !darwin

package mcpserver

import (
	"errors"
	"net"
	"os"
	"time"
)

var errPeerUnsupported = errors.New("reading the peer process is not supported on this platform")

func peerCredentials(conn *net.UnixConn) (pid int, uid int, err error) {
	return 0, 0, errPeerUnsupported
}

func processInfo(pid int) (ProcessInfo, error) {
	return ProcessInfo{PID: pid}, errPeerUnsupported
}

func peerClosed(conn *net.UnixConn) bool {
	return false
}

func waitReadable(f *os.File, timeout time.Duration) (bool, error) {
	return false, errWaitUnsupported
}
