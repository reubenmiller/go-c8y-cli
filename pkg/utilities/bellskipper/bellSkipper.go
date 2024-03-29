package bellskipper

import (
	"io"
)

// BellSkipper implements an io.WriteCloser that skips the terminal bell
// character (ASCII code 7), and writes the rest to os.Stderr. It is used to
// replace readline.Stdout, that is the package used by promptui to display the
// prompts.
//
// This is a workaround for the bell issue documented in
// https://github.com/manifoldco/promptui/issues/49.
type BellSkipper struct {
	w io.WriteCloser
}

func NewBellSkipper(w io.WriteCloser) *BellSkipper {
	return &BellSkipper{
		w: w,
	}
}

// Write implements an io.WriterCloser over os.Stderr, but it skips the terminal
// bell character.
func (bs *BellSkipper) Write(b []byte) (int, error) {
	const charBell = 7 // c.f. readline.CharBell
	if len(b) == 1 && b[0] == charBell {
		return 0, nil
	}
	return bs.w.Write(b)
}

// Close implements an io.WriterCloser over os.Stderr.
func (bs *BellSkipper) Close() error {
	return bs.w.Close()
}
