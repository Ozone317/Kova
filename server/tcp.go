package server

import (
	"fmt"
	"net"
	"bufio"
	"io"
)

func RunTCPServer() {
	var connected_clients int64 = 0

	// Listen for incoming connections on port 7379
	listener, err := net.Listen("tcp", ":7379")
	if err != nil {
		fmt.Println("Error starting server: ", err)
		return
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 7379...")

	// Infinite loop to accept incoming connections
	for {
		// (BLOCKING CALL) Accept a new connection 
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		connected_clients++
		fmt.Println("New client connected: ", conn.RemoteAddr(), "Total connected clients: ", connected_clients)


		// Handle the connection
		go handleConnection(conn, &connected_clients)
	}
}

func handleConnection(conn net.Conn, connected_clients *int64) {
	defer conn.Close()

	// Create a buffer to read data from the client. Wraps the connection in a bufio.Reader for easier reading of lines.
	reader := bufio.NewReader(conn)

	// Infinite loop to continuously read messages from the client until the connection is closed
	for {
		// Read a line of input from the client (blocking call)
		message, err := reader.ReadString('\n')
		if err != nil {
			
			(*connected_clients)--
			fmt.Println("Client disconnected: ", conn.RemoteAddr(), "Total connected clients: ", *connected_clients)

			if err == io.EOF {
				break
			}

			fmt.Println("Error reading from client: ", err)
		}

		// Print received message to the console
		fmt.Printf("Received from %s: %s", conn.RemoteAddr(), message)

		// Echo the message back to the client
		_, err = conn.Write([]byte("Echo: " + message))
		if err != nil {
			fmt.Println("Error writing to client: ", err)
			return
		}

	}
}