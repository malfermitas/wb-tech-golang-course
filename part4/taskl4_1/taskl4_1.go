package main

import (
	"fmt"
	"sync"
	"time"
)

func Or(isRecursiveMerge bool, channels ...<-chan any) <-chan any {
	if isRecursiveMerge {
		return orRecursive(channels...)
	}
	return orNonRecursive(channels...)
}

func orRecursive(channels ...<-chan any) <-chan any {
	orDone := make(chan any)
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	go func() {
		defer close(orDone)
		select {
		case <-channels[0]:
		case <-channels[1]:
		case <-orRecursive(append(channels[2:], orDone)...):
		}
	}()

	return orDone
}

func orNonRecursive(channels ...<-chan any) <-chan any {
	if len(channels) == 0 {
		return nil
	}
	if len(channels) == 1 {
		return channels[0]
	}

	commonCh := make(chan any)
	once := sync.Once{}

	for _, c := range channels {
		go func(c <-chan any) {
			select {
			case <-c:
				once.Do(func() {
					close(commonCh)
				})
			case <-commonCh:
				return
			}
		}(c)
	}

	return commonCh
}

func main() {
	chA := make(chan any)
	chB := make(chan any)
	chC := make(chan any)

	orCh := Or(false, chA, chB, chC)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-orCh
		fmt.Println("Or-канал закрыт, main продолжает работу")
	}()

	time.Sleep(3 * time.Second)
	close(chB)

	wg.Wait()
	fmt.Println("Программа завершена")
}
