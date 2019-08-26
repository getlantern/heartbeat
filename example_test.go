package heartbeat

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

func ExampleSendAndCheckHeartbeat() {
	checkInterval := time.Second
	monitor, _ := NewMonitor(checkInterval)
	var fd uintptr
	if runtime.GOOS == "windows" {
		var err error
		fd, err = monitor.TickerAsFd()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fd = monitor.TickerAsFile().Fd()
	}
	// Now the fd should be passed to the child process
	ticker := FromFd(fd)
	defer ticker.Close()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		fmt.Println(monitor.Wait().Error())
		wg.Done()
	}()
	i := 0
	for ; i < 5; i++ {
		time.Sleep(checkInterval - 200*time.Millisecond)
		fmt.Printf("%d.", i)
		ticker.Tick()
	}
	fmt.Println("")
	time.Sleep(checkInterval + 200*time.Millisecond)
	ticker.Tick()
	fmt.Printf("%d.", i)
	wg.Wait()
	// Output:
	// 0.1.2.3.4.
	// Lost heartbeat
	// 5.
}
