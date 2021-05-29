package fakestdin

import (
	"io/ioutil"
	"log"
	"os"
)

type FakeStdin struct {
	Stdin  *os.File
	backup *os.File
}

func (s *FakeStdin) init() {
	tmpfile, err := ioutil.TempFile("", "fakestdin")
	if err != nil {
		log.Fatal(err)
	}
	s.Stdin = tmpfile
}

func (s *FakeStdin) Restore() {
	if s.backup != nil {
		os.Stdin = s.backup
	}
	if s.Stdin != nil {
		os.Remove(s.Stdin.Name())
	}
}

func (s *FakeStdin) Write(content string) {
	if s.Stdin == nil {
		s.init()
	}
	if _, err := s.Stdin.Write([]byte(content)); err != nil {
		log.Fatal(err)
	}
	if _, err := s.Stdin.Seek(0, 0); err != nil {
		log.Fatal(err)
	}
	os.Stdin = s.Stdin
}

func NewStdIn() *FakeStdin {
	return &FakeStdin{
		backup: os.Stdin,
	}
}
