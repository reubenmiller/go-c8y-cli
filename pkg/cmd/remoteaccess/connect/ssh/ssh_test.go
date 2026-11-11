package ssh

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestBuildSSHArgs(t *testing.T) {
	base := []string{"-o", "ServerAliveInterval=120", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}
	batch := []string{"-o", "BatchMode=yes", "-o", "PreferredAuthentications=publickey", "-o", "ConnectTimeout=30"}
	join := func(parts ...[]string) []string {
		out := []string{}
		for _, p := range parts {
			out = append(out, p...)
		}
		return out
	}
	cases := []struct {
		name string
		opts SSHOptions
		want []string
	}{
		{"no user", SSHOptions{Host: "127.0.0.1", Port: "2222"}, join(base, []string{"-p", "2222", "127.0.0.1"})},
		{"user and command", SSHOptions{Host: "127.0.0.1", Port: "2222", User: "root", Command: []string{"systemctl", "status"}},
			join(base, []string{"-p", "2222", "root@127.0.0.1", "--", "systemctl", "status"})},
		{"preferred auth", SSHOptions{Host: "127.0.0.1", Port: "2222", PreferredAuth: "password"},
			join(base, []string{"-o", "PreferredAuthentications=password", "-p", "2222", "127.0.0.1"})},
		{"batch", SSHOptions{Host: "127.0.0.1", Port: "2222", Batch: true, Command: []string{"uptime"}},
			join(base, batch, []string{"-p", "2222", "127.0.0.1", "--", "uptime"})},
		{"batch ignores preferred auth publickey", SSHOptions{Host: "127.0.0.1", Port: "2222", Batch: true, PreferredAuth: "publickey"},
			join(base, batch, []string{"-p", "2222", "127.0.0.1"})},
		{"port forwarding", SSHOptions{Host: "127.0.0.1", Port: "2222", PortForwarding: "1883:127.0.0.1:1883"},
			join(base, []string{"-p", "2222", "127.0.0.1", "-L", "1883:127.0.0.1:1883"})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildSSHArgs(tc.opts); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got=%v\nwant=%v", got, tc.want)
			}
		})
	}
}

func TestIsAuthFailure(t *testing.T) {
	denied := "Warning: Permanently added '[127.0.0.1]:2222' to the list of known hosts.\r\nme@127.0.0.1: Permission denied (publickey,password).\r\n"
	if !IsAuthFailure(255, denied) {
		t.Error("expected an authentication failure")
	}
	if IsAuthFailure(1, denied) {
		t.Error("exit code 1 is from the remote command, not ssh")
	}
	if IsAuthFailure(255, "ssh: connect to host 127.0.0.1 port 2222: Connection refused") {
		t.Error("connection errors are not authentication failures")
	}
}

func TestConnectWithFallback(t *testing.T) {
	authFailed := "me@127.0.0.1: Permission denied (publickey)."
	t.Run("retries with root after an authentication failure", func(t *testing.T) {
		tried := []string{}
		retried := ""
		code, err := ConnectWithFallback([]string{"", FallbackSSHUser}, func(u string) (int, string, error) {
			tried = append(tried, u)
			if u == "" {
				return 255, authFailed, errors.New("exit status 255")
			}
			return 0, "", nil
		}, func(next string) { retried = next })
		if code != 0 || err != nil || !reflect.DeepEqual(tried, []string{"", "root"}) || retried != "root" {
			t.Fatalf("code=%d err=%v tried=%v retried=%q", code, err, tried, retried)
		}
	})
	t.Run("doesn't retry other failures", func(t *testing.T) {
		tried := 0
		code, _ := ConnectWithFallback([]string{"", FallbackSSHUser}, func(u string) (int, string, error) {
			tried++
			return 3, "", errors.New("exit status 3")
		}, nil)
		if tried != 1 || code != 3 {
			t.Fatalf("expected one attempt with the command's exit code, got tried=%d code=%d", tried, code)
		}
	})
	t.Run("returns the last failure", func(t *testing.T) {
		code, err := ConnectWithFallback([]string{"", FallbackSSHUser}, func(u string) (int, string, error) {
			return 255, authFailed, errors.New("exit status 255")
		}, nil)
		if code != 255 || err == nil {
			t.Fatalf("code=%d err=%v", code, err)
		}
	})
	t.Run("doesn't retry an explicit user", func(t *testing.T) {
		tried := 0
		_, _ = ConnectWithFallback([]string{"admin"}, func(u string) (int, string, error) {
			tried++
			return 255, authFailed, nil
		}, nil)
		if tried != 1 {
			t.Fatalf("expected one attempt, got %d", tried)
		}
	})
}

func TestTailWriter(t *testing.T) {
	w := &tailWriter{max: 8}
	_, _ = w.Write([]byte("0123456789"))
	_, _ = w.Write([]byte("ab"))
	if got := w.String(); got != "456789ab" || !strings.HasSuffix(got, "ab") {
		t.Fatalf("got %q", got)
	}
}
