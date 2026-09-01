package config

type LogConfig struct {
	Level     string
	Format    string
	AddSource bool
}

func newLogConfig(appEnv string) (LogConfig, error) {
	logDefault := getLogDefaults(appEnv)
	addSource, err := getEnvBool("LOG_ADD_SOURCE", logDefault.AddSource)
	if err != nil {
		return LogConfig{}, err
	}

	return LogConfig{
		Level:     getEnv("LOG_LEVEL", logDefault.Level),
		Format:    getEnv("LOG_FORMAT", logDefault.Format),
		AddSource: addSource,
	}, nil
}

func getLogDefaults(appEnv string) LogConfig {
	switch appEnv {
	case "local":
		return LogConfig{
			Level:     "debug",
			Format:    "text",
			AddSource: true,
		}
	case "test":
		return LogConfig{
			Level:     "warn",
			Format:    "text",
			AddSource: false,
		}
	case "staging", "production":
		return LogConfig{
			Level:     "info",
			Format:    "json",
			AddSource: false,
		}
	default:
		return LogConfig{
			Level:     "info",
			Format:    "json",
			AddSource: false,
		}
	}
}
