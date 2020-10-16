package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"path/filepath"
	"strings"
)

// TSSTConfig represents typical structure of service config file
type TSSTConfig struct {
	Services  map[string]ServiceConfig  `json:"services"`
	Databases map[string]DatabaseConfig `json:"databases"`
}

// ServiceConfig represents typical structure of service config
type ServiceConfig struct {
	Enabled bool              `json:"enabled"`
	Address string            `json:"address"`
	Port    string            `json:"port"`
	Options map[string]string `json:"options"`
	Services []string         `json:"services"`
}

// DatabaseConfig represents typical structure of database config
type DatabaseConfig struct {
	Address  string            `json:"address"`
	Dbname   string            `json:"dbname"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Options  map[string]string `json:"options"`
}

// ConfigManager is responsible for reading the settings of the YAML config
type ConfigManager struct {
	viper  *viper.Viper
	Config *TSSTConfig
}

// NewManager returns a new ConfigManager reference
func NewManager(basename string) *ConfigManager {
	mgr := ConfigManager{viper.New(), &TSSTConfig{}}

	configName := strings.TrimSuffix(filepath.Base(basename), filepath.Ext(basename))
	configPath := filepath.Dir(basename)

	mgr.viper.SetConfigType("yaml")
	mgr.viper.SetConfigName(configName)
	mgr.viper.AddConfigPath(configPath)
	mgr.viper.AutomaticEnv()

	return &mgr
}

// Load reads the YAML file and preloads the data into the TSSTConfig struct
func (mgr *ConfigManager) Load() error {
	err := mgr.viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("failed to read in config: %s", err)
	}

	fullCfg, err := mgr.Preload()
	if err != nil {
		return fmt.Errorf("failed to preload config: %s", err)
	}
	mgr.Config = fullCfg

	return nil
}

// Preload binds the config data into the TSSTConfig struct
func (mgr *ConfigManager) Preload() (*TSSTConfig, error) {
	var tsCfg TSSTConfig
	if err := mgr.viper.Unmarshal( &tsCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %q", err)
	}
	return &tsCfg, nil
}

// GetPostgresDSN builds postgres connection string from config values
func (mgr *ConfigManager) GetPostgresDSN() (string, error) {
	pgConf, ok := mgr.Config.Databases["postgres"]
	if !ok {
		return "", errors.New("missing postgres config")
	}

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", pgConf.Username, pgConf.Password, pgConf.Address, pgConf.Dbname), nil
}

func (mgr *ConfigManager) GetServiceConfig(name string) (*ServiceConfig, error) {
	cfg, ok := mgr.Config.Services[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s config is missing", name))
	}

	return &cfg, nil
}

// OpenDB opens connection to postgres
func (mgr *ConfigManager) OpenDB() (*gorm.DB, error) {
	dsn, err := mgr.GetPostgresDSN()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}