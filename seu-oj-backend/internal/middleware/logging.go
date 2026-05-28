package middleware

import (
	"encoding/json"
	"log"
	"os"
)

var structuredLogger = log.New(os.Stdout, "", 0)

func writeJSONLog(entry any) {
	payload, err := json.Marshal(entry)
	if err != nil {
		structuredLogger.Printf(`{"event":"log_marshal_error","error":%q}`, err.Error())
		return
	}
	structuredLogger.Println(string(payload))
}
