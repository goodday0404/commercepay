package config

import "time"

const local_db_url = "postgres://commercepay:commercepay@localhost:5432/commercepay"

type DBConfig struct {
	DBUrl             string
	ConnectionTimeOut time.Duration
}

func newDBConfig(appEnv string) (DBConfig, error) {
	dbUrl, err := getDatabaseURL(appEnv)
	if err != nil {
		return DBConfig{}, err
	}

	dbTimeOut, err := getEnvInt("DB_CONNECT_TIMEOUT", 5)
	if err != nil {
		return DBConfig{}, nil
	}

	return DBConfig{
		DBUrl:             dbUrl,
		ConnectionTimeOut: time.Duration(dbTimeOut) * time.Second,
	}, nil
}

func getDatabaseURL(appEnv string) (string, error) {
	switch appEnv {
	case "local":
		return getEnv("DATABASE_URL", local_db_url), nil
	// case "test":
	// case "staging", "production":
	default:
		return requireEnv("DATABASE_URL")
	}
}
