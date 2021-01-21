package csv

import (
	"os"
	"path"

	"github.com/gocarina/gocsv"
)

// Parse parse csv
func Parse(path string, out interface{}) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err := gocsv.UnmarshalFile(file, out); err != nil {
		panic(err)
	}
}

// Output write csv
func Output(filename string, data interface{}) error {
	dir := path.Dir(filename)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	if err := gocsv.MarshalFile(data, file); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}
