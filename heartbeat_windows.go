package heartbeat

import (
	"os"
	"syscall"
)

func (c *Monitor) TickerAsFile() *os.File {
	panic("not supported")
}

// TickerAsFd returns a file descriptor which can be passed to the child
// process which can reconstruct a ticker from it by calling FromFd.
// For Windows only.
func (c *Monitor) TickerAsFd() (uintptr, error) {
	fd, err := setInheritable(c.fd)
	if err != nil {
		return 0, err
	}
	c.fd = fd
	return fd, err
}

func setInheritable(fd uintptr) (uintptr, error) {
	self, err := syscall.GetCurrentProcess()
	if err != nil {
		return 0, err
	}
	var dup syscall.Handle
	err = syscall.DuplicateHandle(self, syscall.Handle(fd), self, &dup, 0 /*ignored*/, true, syscall.DUPLICATE_SAME_ACCESS)
	if err != nil {
		return 0, err
	}
	syscall.CloseHandle(syscall.Handle(fd))
	return uintptr(dup), nil
}
