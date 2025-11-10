package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

func readFromServer(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			fmt.Println("Reader exited")
			return
		default:
			message := scanner.Text()
			fmt.Println("Message from server --->", message)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Read error:", err)
	}
}

func writeToServer(ctx context.Context, conn net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			fmt.Println("Writer exited")
			return
		default:
			message := scanner.Text()
			_, err := conn.Write([]byte(message + "\n"))
			if err != nil {
				fmt.Println("Write error:", err)
				return
			}
		}
	}
}

func main() {
	fmt.Println("Client started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}

	go readFromServer(ctx, conn)
	go writeToServer(ctx, conn)

	<-stop
	cancel()
	time.Sleep(5 * time.Second)
	fmt.Println("Client stopped")
}
