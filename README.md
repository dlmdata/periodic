# periodic
Perform tasks at regular intervals

## install

	go get github.com/dlmdata/periodic

## usage

Below is an example that shows the common use cases for periodic.

```golang
package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dlmdata/periodic"
)

func sayHello() error {
	fmt.Println("hello world")
	return nil
}

func sayGoodbyeThreeTimes() error {
	if times > 2 {
		return errors.New("too many times")
	}
	fmt.Println("goodbye world")
	times++
	return nil
}

var times = 0

func main() {
	// run a task in the background and ignore errors
	periodic.Background(200*time.Millisecond, periodic.TaskFunc(sayHello))

	// gracefully stop the task at program exit
	task := periodic.NewTask(500*time.Millisecond, periodic.TaskFunc(sayGoodbyeThreeTimes))
	defer task.Stop()

	// log errors from background task
	go func() {
		for err := range task.Error {
			if err != nil {
				log.Println("error:", err)
			}
		}
	}()

	time.Sleep(3 * time.Second)
}
```
