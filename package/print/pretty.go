package print

import (
	"encoding/json"
	"fmt"
	"os"
)

// PrettyPrint prints any struct/interface in a formatted JSON way
func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error pretty printing: %v\n", err)
		fmt.Printf("Raw value: %+v\n", v)
		return
	}
	fmt.Println(string(b))
}

// PrettyPrintToFile writes the pretty printed JSON to a file
func PrettyPrintToFile(v interface{}, filename string) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(b)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

// PrettyPrintString returns the pretty printed JSON as a string
func PrettyPrintString(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling: %w", err)
	}
	return string(b), nil
}
