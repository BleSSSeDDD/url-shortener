package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

func readFromServer(ctx context.Context, conn net.Conn) {
	var res string

	defer conn.Close()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Reader exited")
			return
		default:
			buff := make([]byte, 1024)
			n, _ := conn.Read(buff)
			res += string(buff)
			if n < len(buff) {
				break
			}
			fmt.Println("Message from server --->", string(res))
		}
	}
}

func writeToServer(ctx context.Context, conn net.Conn) {
	var str string
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Writer exited")
			return
		default:
			fmt.Println("Input your message to server: ")
			fmt.Scanln(&str)
			conn.Write([]byte(str + "\n"))
		}
	}
}

func main() {
	fmt.Println("Client started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err1 := net.Dial("tcp", "127.0.0.1:8080")
	if err1 != nil {
		fmt.Println(err1)
		return
	}

	go readFromServer(ctx, conn)
	go writeToServer(ctx, conn)

	<-stop
	cancel()
	time.Sleep(10 * time.Second)
	fmt.Println("Client stopped")

}
