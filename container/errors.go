package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type errorDetail struct {
	Message string `json:"message"`
}

type message struct {
	Stream      string       `json:"stream"`
	Error       string       `json:"error"`
	ErrorDetail *errorDetail `json:"errorDetail"`
}

func readErr(r io.Reader) error {
	dec := json.NewDecoder(r)
	for {
		var msg message

		if err := dec.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if msg.ErrorDetail != nil {
			return fmt.Errorf("%s", msg.ErrorDetail.Message)
		}

		if msg.Error != "" {
			return fmt.Errorf("%s", msg.Error)
		}
	}

}
