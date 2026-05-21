package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Project      string                 `mapstructure:"project"`
	Version      string                 `mapstructure:"version"`
	Registry     string                 `mapstructure:"registry"`
	Environments map[string]Environment `mapstructure:"environments"`
	Deploy       DeployDefaults         `mapstructure:"deploy"`
}

type Environment struct {
	Host       string            `mapstructure:"host"`
	SSHUser    string            `mapstructure:"ssh_user"`
	SSHPort    int               `mapstructure:"ssh_port"`
	DeployPath string            `mapstructure:"deploy_path"`
	Services   []Service         `mapstructure:"services"`
	EnvVars    map[string]string `mapstructure:"env_vars"`
}

type Service struct {
	Name      string `mapstructure:"name"`
	Port      int    `mapstructure:"port"`
	HealthURL string `mapstructure:"health_url"`
	Container string `mapstructure:"container"`
	Replicas  int    `mapstructure:"replicas"`
	Image     string `mapstructure:"image"`
}

type DeployDefaults struct {
	Strategy string `mapstructure:"strategy"`
	Timeout  int    `mapstructure:"timeout"`
	Retries  int    `mapstructure:"retries"`
}

func Load(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".forge")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
	}

	viper.SetEnvPrefix("FORGE")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	setDefaults(cfg)
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Deploy.Strategy == "" {
		cfg.Deploy.Strategy = "rolling"
	}
	if cfg.Deploy.Timeout == 0 {
		cfg.Deploy.Timeout = 300
	}
	if cfg.Deploy.Retries == 0 {
		cfg.Deploy.Retries = 3
	}
	for name, env := range cfg.Environments {
		if env.SSHPort == 0 {
			env.SSHPort = 22
		}
		for i := range env.Services {
			if env.Services[i].Replicas == 0 {
				env.Services[i].Replicas = 1
			}
		}
		cfg.Environments[name] = env
	}
}

func (c *Config) GetEnvironment(name string) (Environment, error) {
	env, ok := c.Environments[name]
	if !ok {
		available := make([]string, 0, len(c.Environments))
		for k := range c.Environments {
			available = append(available, k)
		}
		return Environment{}, fmt.Errorf("environment %q not found (available: %s)", name, strings.Join(available, ", "))
	}
	return env, nil
}

func (e *Environment) GetService(name string) (Service, error) {
	for _, s := range e.Services {
		if s.Name == name {
			return s, nil
		}
	}
	available := make([]string, 0, len(e.Services))
	for _, s := range e.Services {
		available = append(available, s.Name)
	}
	return Service{}, fmt.Errorf("service %q not found (available: %s)", name, strings.Join(available, ", "))
}

func (c *Config) EnvNames() []string {
	names := make([]string, 0, len(c.Environments))
	for k := range c.Environments {
		names = append(names, k)
	}
	return names
}
