package heartbeat

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func helperProcess(params []string, extraFiles []*os.File) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, params...)
	env := []string{
		"GO_WANT_HELPER_PROCESS=1",
	}

	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(env, os.Environ()...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.ExtraFiles = extraFiles
	return cmd
}

// This is executed by `helperProcess` in a separate process in order to
// provider a proper sub-process environment to test some of our functionality.
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Find the arguments to our helper, which are the arguments past
	// the "--" in the command line.
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fd := fs.Uint("fd", 0, "")
	i1 := fs.Duration("i1", 0, "")
	count := fs.Int("c", 0, "")
	i2 := fs.Duration("i2", 0, "")
	fs.Parse(args)
	ticker := FromFd(uintptr(*fd))
	for i := 0; ; i++ {
		if i < *count {
			time.Sleep(*i1)
		} else {
			time.Sleep(*i2)
		}
		fmt.Printf("%d.", i)
		fmt.Fprintf(os.Stderr, "%d.", i)
		if err := ticker.Tick(); err != nil {
			fmt.Fprintf(os.Stderr, "Tick failed: %v\n", err)
		}
	}
}

func TestSendAndCheckHeartbeat(t *testing.T) {
	stdout := new(bytes.Buffer)
	monitor, _ := NewMonitor(time.Second)
	extraFiles := []*os.File{}
	var fd uintptr
	if runtime.GOOS == "windows" {
		var err error
		fd, err = monitor.TickerAsFd()
		if err != nil {
			t.Fatal(err)
		}
	} else {
		// 3 corresponds to the first item in os.Cmd.ExtraFiles
		fd = 3
		extraFiles = []*os.File{monitor.TickerAsFile()}
	}
	// sends heartbeat every 800ms for 10 times, then every 2s.
	params := []string{
		"-fd", strconv.Itoa(int(fd)),
		"-i1", "800ms",
		"-c", "10",
		"-i2", "2s",
	}
	p := helperProcess(params, extraFiles)
	p.Stdout = stdout
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	err := monitor.Wait()
	if err != Lost {
		t.Errorf("Unexpected Wait error %v", err)
	}
	p.Process.Kill()
	if stdout.String() != "0.1.2.3.4.5.6.7.8.9." {
		t.Errorf("Should have received exactly 10 heartbeats, got %s", stdout.String())
	}
}
