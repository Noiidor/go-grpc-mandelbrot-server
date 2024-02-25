package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// ListenSignals when the function receives a signal, it stops recordings signals
// and displays a message in the log indicating which signal was received.
func ListenSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sigVal := <-sigChan
	signal.Stop(sigChan)
	fmt.Printf("stop signal: %v", sigVal)
}
