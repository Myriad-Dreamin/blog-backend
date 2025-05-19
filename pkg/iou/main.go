package iou

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

func WriteJsonToFile(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "error create file")
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return errors.Wrap(err, "error encode data")
	}

	return nil
}
