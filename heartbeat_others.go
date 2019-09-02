// +build !windows

package heartbeat

import "os"

// TickerAsFile returns an os.File to be passed to the child process via the
// ExtraFiles field in os.Cmd struct, which can later be reconstructed by
// calling FromFd by the corresponding file descriptor in the child process.
// On Windows, use TickerAsFd instead.
func (c *Monitor) TickerAsFile() *os.File {
	return os.NewFile(c.fd, "heartbeat-ticker")
}

func (c *Monitor) TickerAsFd() (uintptr, error) {
	panic("not supported")
}
