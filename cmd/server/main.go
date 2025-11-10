package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Hello from TCP server!\n"))

	for {
		select {
		case <-ctx.Done():
			log.Println("Connection stopped by context")
			return
		default:
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				return
			}

			message := string(buf[:n])
			log.Printf("Received: %s", message)

			response := "Your message was received, its \"" + message + "\"\n"
			if _, err := conn.Write([]byte(response)); err != nil {
				return
			}
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				tcpConn.SetNoDelay(true)
			}
		}
	}
}

func startServer(ctx context.Context) error {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConnection(ctx, conn)
	}
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		fmt.Println("Server started")
		if startError := startServer(ctx); startError != nil {
			fmt.Println(startError)
		}
	}()

	<-stop
	fmt.Println("Graceful shutdown")
}
