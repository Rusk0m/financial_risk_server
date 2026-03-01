package config

// Config представляет полную конфигурацию приложения
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logger   LoggerConfig   `yaml:"logger"`
}

// ServerConfig содержит настройки HTTP сервера
type ServerConfig struct {
	Port         string `yaml:"port"`
	Environment  string `yaml:"environment"`
	ReadTimeout  int    `yaml:"read_timeout"`  // в секундах
	WriteTimeout int    `yaml:"write_timeout"` // в секундах
	IdleTimeout  int    `yaml:"idle_timeout"`  // в секундах
}

// DatabaseConfig содержит настройки подключения к базе данных
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"sslmode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // в секундах
}

// LoggerConfig содержит настройки логирования
type LoggerConfig struct {
	Level    string `yaml:"level"`  // debug, info, warn, error
	Format   string `yaml:"format"` // json, text
	Output   string `yaml:"output"` // stdout, file
	FilePath string `yaml:"file_path"`
}
