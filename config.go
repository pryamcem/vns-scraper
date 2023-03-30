package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	ErrFileNotExists = errors.New("File with this path is not exists")
)

type Configuration struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func ReadConfig(path string) (error, Configuration) {
	config := Configuration{}
	jsonFile, err := os.Open(path)
	if err != nil {
		return ErrFileNotExists, config
	}
	defer jsonFile.Close()

	byteFile, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("Reading file to bytes error: %w", err), config
	}

	err = json.Unmarshal(byteFile, &config)
	if err != nil {
		return fmt.Errorf("Unmarshaling error: %w", err), config
	}

	return nil, config
}

func NewConfig(path string, config Configuration) error {
	fmt.Println(config)
	content, err := json.MarshalIndent(config, "", " ")
	fmt.Println("Content", string(content))
	if err != nil {
		return fmt.Errorf("Can't marshal config to Json: %w", err)
	}
	err = ioutil.WriteFile(path, content, 0644)
	if err != nil {
		return fmt.Errorf("Can't write to Json file: %w", err)
	}
	return nil
}
