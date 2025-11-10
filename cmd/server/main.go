package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func getLinesFromChannel(f io.ReadCloser) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		tmp := make([]byte, 8)

		for {
			n, err := f.Read(tmp)
			if n > 0 {
				out <- string(tmp[:n])
				time.Sleep(100 * time.Millisecond)
			}
			if err != nil {
				break
			}
		}
	}()

	return out
}
func main() {
	f, err := os.Open("cmd/server/test.txt")
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		for str := range getLinesFromChannel(f) {
			fmt.Println(str)
		}
	}()
	time.Sleep(2 * time.Second)

}
