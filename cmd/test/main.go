package main

import (
	"fmt"
	"time"
)

func main() {
	a := 1

	go func() {
		for i := 0; i <= 100; i++ {
			a = i
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for true {
			if a % 5 != 0 {
				fmt.Println(a)
				time.Sleep(2 * time.Second)
			}
		}
	}()

	time.Sleep(600 * time.Second)
}