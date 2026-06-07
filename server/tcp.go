package server

import (
	"fmt"
	"io"
	"net"

	"github.com/Ozone317/Kova/core"
)

func readCommand(conn io.ReadWriter) (*core.KovaCmd, error) {
	var buffer []byte = make([]byte, 512)
	n, err := conn.Read(buffer[:])
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buffer[:n])
	if err != nil {
		return nil, err
	}

	cmd := core.KovaCmd{
		Cmd:  tokens[0],
		Args: tokens[1:],
	}
	return &cmd, nil
}

func respond(command *core.KovaCmd, conn io.ReadWriter) {
	err := core.EvalAndRespond(command, conn)
	if err != nil {
		respondError(err, conn)
	}
}

func respondError(err error, conn io.ReadWriter) {
	_, err = conn.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
	if err != nil {
		fmt.Println("Error writing to client: ", err)
	}
}

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

	// Infinite loop to continuously read messages from the client until the connection is closed
	for {
		// Read a line of input from the client (blocking call)
		command, err := readCommand(conn)
		if err != nil {
			(*connected_clients)--
			fmt.Println("Client disconnected: ", conn.RemoteAddr(), "Total connected clients: ", *connected_clients)

			if err == io.EOF {
				break
			}

			fmt.Println("Error reading from client: ", err)
			continue
		}

		respond(command, conn)

	}
}
