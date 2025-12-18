package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/rs/zerolog/log"
)

type AppConfig struct {
	App struct {
		Name        string `yaml:"name" json:"name"`
		Environment string `yaml:"environment" json:"environment"`
	} `yaml:"app" json:"app"`

	Server struct {
		Host string `yaml:"host" json:"host"`
		Port int    `yaml:"port" json:"port"`
	} `yaml:"server" json:"server"`

	Logging struct {
		Level  string `yaml:"level" json:"level"`
		Output string `yaml:"output" json:"output"`
	} `yaml:"logging" json:"logging"`

	Databases []struct {
		Name   string `yaml:"name" json:"name"`
		Engine string `yaml:"engine" json:"engine"`
		URI    string `yaml:"uri" json:"uri"`
		Pool   struct {
			Max int `yaml:"max" json:"max"`
			Min int `yaml:"min" json:"min"`
		} `yaml:"pool" json:"pool"`
	} `yaml:"databases" json:"databases"`

	Modules []Module `yaml:"modules" json:"modules"`

	Auths []Auth `yaml:"auths" json:"auths"`
}

type Auth struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`

	JWTPubKey64   string `yaml:"jwt_pubkey64,omitempty" json:"jwt_pubkey64,omitempty"`
	JWTPrivKey64  string `yaml:"jwt_privkey64,omitempty" json:"jwt_privkey64,omitempty"`
	BasicUsername string `yaml:"basic_username,omitempty" json:"basic_username,omitempty"`
	BasicPassword string `yaml:"basic_password,omitempty" json:"basic_password,omitempty"`
}

type Module struct {
	Name       string   `yaml:"name" json:"name"`
	Database   string   `yaml:"database" json:"database"`
	Table      string   `yaml:"table" json:"table"`
	Auth       string   `yaml:"auth,omitempty" json:"auth,omitempty"`
	Fields     []string `yaml:"fields" json:"fields"`
	Operations []string `yaml:"operations" json:"operations"`
}

// -------------------------------------
// OPTIONAL: auto-detect recipe file
// -------------------------------------

func detectConfigFile() string {
	if _, err := os.Stat("recipe.yaml"); err == nil {
		return "recipe.yaml"
	}
	if _, err := os.Stat("recipe.yml"); err == nil {
		return "recipe.yml"
	}
	if _, err := os.Stat("recipe.json"); err == nil {
		return "recipe.json"
	}
	return "" // no file found
}

func loadFromFile(path string) (*AppConfig, error) {
	log.Info().Msg("Load Config from File")
	if path == "" {
		return &AppConfig{}, nil // no file, empty recipe
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &cfg)
	case ".json":
		err = json.Unmarshal(data, &cfg)
	default:
		return nil, errors.New("unsupported recipe format")
	}

	return &cfg, err
}

// -------------------------------------

func loadFromEnvConfig() (*AppConfig, error) {
	log.Info().Msg("Load Config from Environment")
	y := os.Getenv("CONFIG_YAML")
	j := os.Getenv("CONFIG_JSON")

	if y == "" && j == "" {
		return nil, nil
	}

	var cfg AppConfig
	var err error

	if y != "" {
		err = yaml.Unmarshal([]byte(y), &cfg)
	} else {
		err = json.Unmarshal([]byte(j), &cfg)
	}

	return &cfg, err
}

// -------------------------------------

func applyEnvOverrides(cfg *AppConfig) {
	log.Info().Msg("Override Config from Environment")
	if v := os.Getenv("APP_NAME"); v != "" {
		cfg.App.Name = v
	}
	if v := os.Getenv("APP_ENVIRONMENT"); v != "" {
		cfg.App.Environment = v
	}
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = n
		}
	}
}

// -------------------------------------

func merge(dst, src *AppConfig) {
	// strings
	if src.App.Name != "" {
		dst.App.Name = src.App.Name
	}
	if src.App.Environment != "" {
		dst.App.Environment = src.App.Environment
	}
	if src.Server.Host != "" {
		dst.Server.Host = src.Server.Host
	}
	if src.Server.Port != 0 {
		dst.Server.Port = src.Server.Port
	}

	if src.Logging.Level != "" {
		dst.Logging.Level = src.Logging.Level
	}
	if src.Logging.Output != "" {
		dst.Logging.Output = src.Logging.Output
	}

	// slices: full replace
	if len(src.Databases) > 0 {
		dst.Databases = src.Databases
	}
	if len(src.Modules) > 0 {
		dst.Modules = src.Modules
	}
	if len(src.Auths) > 0 {
		dst.Auths = src.Auths
	}
}

// ------------------------
// Auto loader (file + env + overrides)
// ------------------------

func LoadAuto(paths ...string) (*AppConfig, error) {
	cfg := &AppConfig{}

	var path string
	if len(paths) > 0 && paths[0] != "" {
		path = paths[0] // user provided recipe path
	} else {
		path = detectConfigFile() // auto-detect recipe file
	}

	// 1. Load from file if path available
	fileCfg, err := loadFromFile(path)
	if err != nil {
		return nil, err
	}
	merge(cfg, fileCfg)

	// 2. Load from ENV (CONFIG_YAML / CONFIG_JSON)
	envCfg, err := loadFromEnvConfig()
	if err != nil {
		return nil, err
	}
	if envCfg != nil {
		merge(cfg, envCfg)
	}

	// 3. ENV overrides (highest priority)
	applyEnvOverrides(cfg)

	return cfg, nil
}

// Find DB Engine by Name
func GetDBEngineByName(cfg *AppConfig, name string) string {
	for _, db := range cfg.Databases {
		if db.Name == name {
			return db.Engine
		}
	}
	return "" // not found
}

// Find Auth by Name
func FindAuth(cfg *AppConfig, name string) *Auth {
	for _, a := range cfg.Auths {
		if a.Name == name {
			return &a
		}
	}
	return nil
}
