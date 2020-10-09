package config

import (
	"io/ioutil"
)

// TraceReplayerConfig contains necessary auth info for trace replayer
type TraceReplayerConfig struct {
	Username   string
	Password   string
	Repository string
	URL        string
}

// SetTraceReplayerEnvVariables sets env variables for the trace replayer to authenticate with registry
func SetTraceReplayerEnvVariables(filepath string, traceReplayerConfig TraceReplayerConfig) error {
	envVariables := []byte("REGISTRY_USERNAME=" + traceReplayerConfig.Username + "\nREGISTRY_PASSWORD=" + traceReplayerConfig.Password + "\nREGISTRY_REPO=" + traceReplayerConfig.Repository + "\nREGISTRY_URL=" + traceReplayerConfig.URL + "\n")
	err := ioutil.WriteFile(filepath+"/.env", envVariables, 0644)
	if err != nil {
		return err
	}
	return nil
}
