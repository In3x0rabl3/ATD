package utils

import (
	"fmt"
	"log"
	"os"
)

func LogFlagSubmission(userID, module, flag string) {
	logFilePath := "./data/flags/correct_flags.log"
	logEntry := fmt.Sprintf("%s submitted correct flag for %s: %s\n", userID, module, flag)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error logging flag: %v", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(logEntry); err != nil {
		log.Printf("Error writing to flag log file: %v", err)
	}
}
