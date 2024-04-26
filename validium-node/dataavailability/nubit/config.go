package nubit

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	RpcURL    string `json:"rpcURL,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	AuthKey   string `json:"authKey,omitempty"`
}

func (c *Config) GetConfig(configFileName string) error {
	jsonFile, err := os.Open(configFileName)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, &c)
	if err != nil {
		return err
	}
	return nil
}
