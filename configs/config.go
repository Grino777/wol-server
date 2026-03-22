package configs

import "os"

// Prod/Dev mode
const (
	LogMode  = "development"
	LogLevel = "debug" // debug, info, warn, error
)

// Max request body size for wake requests (64Kb)
const (
	MaxWakeRequestBodyBytes = 64_000
)

var UserPassword = os.Getenv("USER_PASSWORD")
var Username = os.Getenv("USERNAME")
