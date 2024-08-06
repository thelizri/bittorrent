package utils

import (
	"fmt"
	"log"
	"os"
)

const LINE_SEPARATOR string = "-------------------------------------------------------------------------"

func LogSeparator() {
	log.Printf("\n%s\n", LINE_SEPARATOR)
	fmt.Printf("\n%s\n", LINE_SEPARATOR)
}

func LogAndPrintln(message string) {
	log.Println(message)
	fmt.Println(message)
}

func LogAndPrintf(format string, args ...interface{}) {
	// Use fmt.Sprintf to format the string with the provided arguments
	message := fmt.Sprintf(format, args...)

	// Check if the message ends with a newline character
	if len(message) == 0 || message[len(message)-1] != '\n' {
		message += "\n" // Append a newline character if missing
	}

	// Print the formatted message to the console
	fmt.Print(message)

	// Log the formatted message
	log.Print(message)
}

func LogPrintExit(message string) {
	log.Println(message)
	fmt.Println(message)
	os.Exit(1)
}
