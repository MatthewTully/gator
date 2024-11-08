package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func getConfigFilePath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("an Error occurred getting Home Path: %v", err)
	}
	return filepath.Join(path, configFileName), nil
}

func write(c Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	file_data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("an Error occurred Marshalling JSON data: %v", err)
	}
	err = os.WriteFile(path, file_data, 0666)
	if err != nil {
		return fmt.Errorf("an Error occurred writing to file: %v", err)
	}
	return nil
}

func Read() Config {
	path, err := getConfigFilePath()
	if err != nil {
		fmt.Println(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(fmt.Errorf("an Error occurred Reading config file: %v", err))
	}
	var conf Config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		fmt.Println(fmt.Errorf("an Error occurred Unmarshalling JSON data: %v", err))
	}
	return conf

}
