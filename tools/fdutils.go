package tools

import (
	"os"
	"io/ioutil"
)

func FSExists(dir string) bool {
	if _, err := os.Stat(dir); err != nil {
		return false
	}
	return true
}

func CreateDirectory(dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func WriteToFile(dir string, text []byte) error {
	err := ioutil.WriteFile(dir, text, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func ReadFromFile(dir string) ([]byte, error) {
	b, err := ioutil.ReadFile(dir)
	if err != nil {
		return nil, err
	}
	return b, nil
}