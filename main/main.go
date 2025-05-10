package main

//go:generate go run ./gen/gen.go

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pmeier/redgiant/internal/errors"
	"github.com/rs/zerolog"
)

// import "github.com/pmeier/redgiant/internal/cmd"

// func main() {
// 	cmd.Execute()
// }

func main() {
	var err error = errors.New("oops").HTTPStatusCode(http.StatusBadGateway).HTTPRedacted(false)
	// var err error = nerrors.New("heheheh")

	logger := zerolog.New(os.Stdout)

	var code int
	var logData map[string]any
	var httpData map[string]any
	if rgerr, ok := err.(*errors.Error); ok {
		code = rgerr.StatusCode
		if err := json.Unmarshal([]byte(err.Error()), &logData); err != nil {
			panic(err.Error())
		}
		if !rgerr.Redacted {
			httpData = logData
		}
	} else {
		code = http.StatusInternalServerError
	}

	if httpData == nil {
		httpData = map[string]any{zerolog.MessageFieldName: http.StatusText(code)}
	}

	e := logger.Error()
	for k, v := range logData {
		e.Any(k, v)
	}
	e.Send()

	if v, err := json.MarshalIndent(httpData, "", "  "); err != nil {
		panic(err.Error())
	} else {
		fmt.Println(code, string(v))
	}

}
