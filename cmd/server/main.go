package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Open("cmd/server/test.txt")
	if err != nil {
		fmt.Println(err)
	}
	res := make([]byte, 5)
	for {
		n, err := f.Read(res)
		fmt.Printf("[%s]", res[:n])
		if n < len(res) {
			fmt.Println(err)
			break
		}
	}

}
