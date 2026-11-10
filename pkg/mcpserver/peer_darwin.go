//go:build darwin

package mcpserver

import (
	"encoding/binary"
	"net"

	"golang.org/x/sys/unix"
)

func peerCredentials(conn *net.UnixConn) (pid int, uid int, err error) {
	raw, err := conn.SyscallConn()
	if err != nil {
		return 0, 0, err
	}
	var cred *unix.Xucred
	var credErr error
	if err := raw.Control(func(fd uintptr) {
		cred, credErr = unix.GetsockoptXucred(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
		if credErr == nil {
			pid, credErr = unix.GetsockoptInt(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERPID)
		}
	}); err != nil {
		return 0, 0, err
	}
	if credErr != nil {
		return 0, 0, credErr
	}
	return pid, int(cred.Uid), nil
}

func processInfo(pid int) (ProcessInfo, error) {
	info := ProcessInfo{PID: pid}
	kinfo, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil {
		return info, err
	}
	info.PPID = int(kinfo.Eproc.Ppid)
	info.StartTime = kinfo.Proc.P_starttime.Sec*1_000_000 + int64(kinfo.Proc.P_starttime.Usec)
	info.Name = unix.ByteSliceToString(kinfo.Proc.P_comm[:])

	// kern.procargs2: argc (int32), executable path, padding, then the arguments
	if buf, err := unix.SysctlRaw("kern.procargs2", pid); err == nil && len(buf) > 4 {
		argc := int(binary.LittleEndian.Uint32(buf[:4]))
		parts := splitNul(buf[4:])
		if len(parts) > 0 {
			info.Executable = parts[0]
			if len(parts) > argc {
				info.Args = parts[1 : argc+1]
			}
		}
	}
	return info, nil
}
