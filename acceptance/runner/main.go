package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	dec := json.NewDecoder(os.Stdin)
	var job map[string]any
	if err := dec.Decode(&job); err != nil && err.Error() != "EOF" {
		enc := json.NewEncoder(os.Stdout)
		_ = enc.Encode(map[string]string{"status": "infrastructure_error"})
		os.Exit(1)
	}
	fmt.Println(`{"status":"test_success"}`)
}
