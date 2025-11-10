package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	conn.Write([]byte("Hello from TCP server!\n"))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Println("Connection stopped by context")
			return
		default:
			message := scanner.Text()
			log.Printf("Received: %s", message)

			response := "Your message was received: \"" + message + "\"\n"
			if _, err := conn.Write([]byte(response)); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Read error: %v", err)
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
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
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
		fmt.Println("Server started on :8080")
		if startError := startServer(ctx); startError != nil {
			fmt.Println(startError)
		}
	}()

	<-stop
	fmt.Println("Graceful shutdown")
	cancel()
}
