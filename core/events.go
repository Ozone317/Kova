package core

import "fmt"

func Shutdown() {
	fmt.Println("\nShutting down the server...")
	_, err := evalBGREWRITEAOF([]string{})
	if err != nil {
		fmt.Println("Error rewriting AOF file:", err)
	}
	fmt.Println("Server shut down successfully")
}
