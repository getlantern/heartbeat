package heartbeat

import (
	"errors"
	"os"
	"time"
)

var (
	Lost = errors.New("Lost heartbeat")
)

type Monitor struct {
	r             *os.File
	fd            uintptr
	checkInterval time.Duration
}

func NewMonitor(checkInterval time.Duration) (*Monitor, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	return &Monitor{r, w.Fd(), checkInterval}, nil
}

// Wait keeps monitoring heartbeat from the ticker in the configured interval.
// It returns heartbeat.Lost when no heartbeat detected in time, or whichever
// error receiving data from the ticker.
func (c *Monitor) Wait() error {
	chHeartbeat := make(chan error)
	go func() {
		buf := make([]byte, 1)
		defer c.r.Close()
		for {
			_, err := c.r.Read(buf)
			chHeartbeat <- err
			if err != nil {
				return
			}
		}
	}()

	tk := time.NewTimer(c.checkInterval)
	defer tk.Stop()
	for {
		select {
		case err := <-chHeartbeat:
			if err != nil {
				return err
			}
			// got heartbeat, reset timer
			if !tk.Stop() {
				<-tk.C
			}
			tk.Reset(c.checkInterval)
		case <-tk.C:
			return Lost
		}
	}
}

type Ticker interface {
	// Tick sends a heartbeat to the corresponding checker
	Tick() error
	Close() error
}

type ticker struct {
	w *os.File
	b []byte
}

// FromFd reconstructs a ticker from a file descriptor, typically passed from
// the parent process.
func FromFd(fd uintptr) Ticker {
	f := os.NewFile(fd, "heartbeat-ticker")
	return &ticker{f, make([]byte, 1)}
}

func (s *ticker) Tick() error {
	_, err := s.w.Write(s.b)
	return err
}

func (s *ticker) Close() error {
	return s.w.Close()
}
