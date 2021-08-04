package crawlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type Config struct {
	Districts []string `json:"kerületek"`
	MinPrice  int      `json:"min_ár"`
	MaxPrice  int      `json:"max_ár"`
	MinSize   int      `json:"min_méret"`
	MaxSize   int      `json:"max_méret"`
	Type      string   `json:"lakás_vagy_ház"`
}

func ReadJsonConfig(configfile string) (Config, error) {
	file, err := ioutil.ReadFile(configfile)
	if err != nil {
		return Config{}, err
	}

	var config Config

	err = json.Unmarshal([]byte(file), &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func CreateFileNameFromConfig(c Config, prefix string) string {
	if !strings.HasSuffix(prefix, "_") && len(prefix) != 0 {
		prefix += "_"
	}
	districts := strings.Join(c.Districts, "_")

	return fmt.Sprintf("%sar_%d_%d_meret_%d_%d_kerulet_%s_%s.csv", prefix, c.MinPrice, c.MaxPrice, c.MinSize, c.MaxSize, districts, c.Type)
}
