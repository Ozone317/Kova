package core

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Ozone317/Kova/config"
)

func dumpKey(file *os.File, key string, obj *Object) {
	command := fmt.Sprintf("SET %s %s", key, obj.value)
	tokens := strings.Split(command, " ")
	b := Encode(tokens, false)
	_, err := file.Write(b)
	if err != nil {
		log.Printf("Error writing to AOF file: %v", err)
	}
}

// TODO: make this a background job (by forking a process)
func DumpAllAOF() {
	file, err := os.OpenFile(config.AOF_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println("Error opening AOF file:", err)
		return
	}
	defer file.Close()

	log.Printf("rewriting AOF file at: %s", config.AOF_FILE)

	for key, obj := range store {
		dumpKey(file, key, obj)
	}

	log.Printf("AOF file rewritten successfully")
}
