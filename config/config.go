package config

type Config struct {
	ServerPort string
	DBPath     string
}

func Load() *Config {
	return &Config{
		ServerPort: ":8080",
		DBPath:     "escape_room.db",
	}
}
