//go:build linux

package mcpserver

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func peerCredentials(conn *net.UnixConn) (pid int, uid int, err error) {
	raw, err := conn.SyscallConn()
	if err != nil {
		return 0, 0, err
	}
	var cred *unix.Ucred
	var credErr error
	if err := raw.Control(func(fd uintptr) {
		cred, credErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	}); err != nil {
		return 0, 0, err
	}
	if credErr != nil {
		return 0, 0, credErr
	}
	return int(cred.Pid), int(cred.Uid), nil
}

func processInfo(pid int) (ProcessInfo, error) {
	info := ProcessInfo{PID: pid}
	procDir := filepath.Join("/proc", strconv.Itoa(pid))

	stat, err := os.ReadFile(filepath.Join(procDir, "stat"))
	if err != nil {
		return info, err
	}
	// Format: pid (comm) state ppid ... where comm can contain spaces and brackets
	end := strings.LastIndexByte(string(stat), ')')
	start := strings.IndexByte(string(stat), '(')
	if start < 0 || end < start {
		return info, fmt.Errorf("unexpected format of %s/stat", procDir)
	}
	info.Name = string(stat[start+1 : end])
	fields := strings.Fields(string(stat[end+1:]))
	if len(fields) >= 2 {
		info.PPID, _ = strconv.Atoi(fields[1])
	}
	// field 22 (starttime), counted from field 3 (state) which follows the command name
	if len(fields) >= 20 {
		info.StartTime, _ = strconv.ParseInt(fields[19], 10, 64)
	}

	if exe, err := os.Readlink(filepath.Join(procDir, "exe")); err == nil {
		info.Executable = exe
	}
	if cmdline, err := os.ReadFile(filepath.Join(procDir, "cmdline")); err == nil {
		info.Args = splitNul(cmdline)
	}
	return info, nil
}
