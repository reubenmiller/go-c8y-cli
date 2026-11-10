//go:build linux || darwin

package mcpserver

import (
	"errors"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// peerClosed returns true if the other end of the connection has closed it.
// It peeks at the socket, so no data is consumed.
func peerClosed(conn *net.UnixConn) bool {
	raw, err := conn.SyscallConn()
	if err != nil {
		return false
	}
	closed := false
	_ = raw.Control(func(fd uintptr) {
		buf := make([]byte, 1)
		n, _, err := unix.Recvfrom(int(fd), buf, unix.MSG_PEEK|unix.MSG_DONTWAIT)
		switch {
		case err == nil && n == 0:
			closed = true
		case err != nil && !errors.Is(err, unix.EAGAIN) && !errors.Is(err, unix.EWOULDBLOCK) && !errors.Is(err, unix.EINTR):
			closed = true
		}
	})
	return closed
}

// waitReadable waits until the file has data to read, or the timeout expires.
// Select is used rather than poll, as poll does not support terminals on macOS.
func waitReadable(f *os.File, timeout time.Duration) (bool, error) {
	fd := int(f.Fd())
	if fd < 0 || fd >= unix.FD_SETSIZE {
		return false, errWaitUnsupported
	}
	var set unix.FdSet
	set.Set(fd)
	tv := unix.NsecToTimeval(timeout.Nanoseconds())
	n, err := unix.Select(fd+1, &set, nil, nil, &tv)
	if errors.Is(err, unix.EINTR) {
		return false, nil
	}
	return n > 0, err
}
