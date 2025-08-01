package ui

import (
	"fmt"
	"time"
)

func WithSpinner(message string, fn func() error) error {
	fmt.Printf("%s ", message)

	done := make(chan bool)
	started := make(chan bool)
	go func() {
		chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		started <- true
		for {
			select {
			case <-done:
				fmt.Print("\r\u001b[K")
				return
			default:
				fmt.Printf("\r%s %s", message, chars[i%len(chars)])
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	<-started
	err := fn()
	done <- true
	time.Sleep(50 * time.Millisecond)

	if err != nil {
		fmt.Printf("\r%s ❌\n", message)
	} else {
		fmt.Printf("\r%s ✅\n", message)
	}

	return err
}
