package config

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// TraceReplayerCredentials contains necessary auth info for trace replayer
type TraceReplayerCredentials struct {
	Username   string
	Password   string
	Repository string
	URL        string
}

// TraceReplayerConfig contains finer grained config for trace replayer
type TraceReplayerConfig struct {
	TracePath     string `yaml:"trace-path,omitempty"`
	Clients       []string
	ClientThreads int      `yaml:"client-threads,omitempty"`
	TraceDir      string   `yaml:"trace-directory,omitempty"`
	TraceFiles    []string `yaml:"trace-files,omitempty"`
	Wait          bool
	WarmupThreads int    `yaml:"warmup-threads,omitempty"`
	MasterPort    int    `yaml:"master-port,omitempty"`
	LimitType     string `yaml:"limit-type,omitempty"`
	LimitAmount   int    `yaml:"limit-amount,omitempty"`
	ResultsDir    string `yaml:"results-directory,omitempty"`
}

// SetTraceReplayerEnvVariables sets env variables for the trace replayer to authenticate with registry
func SetTraceReplayerEnvVariables(traceReplayerCreds TraceReplayerCredentials, traceReplayerConfig TraceReplayerConfig) error {
	envVariables := []byte("REGISTRY_USERNAME=" + traceReplayerCreds.Username +
		"\nREGISTRY_PASSWORD=" + traceReplayerCreds.Password +
		"\nREGISTRY_REPO=" + traceReplayerCreds.Repository +
		"\nREGISTRY_URL=" + traceReplayerCreds.URL +
		"\nCLIENTS=" + strings.Join(traceReplayerConfig.Clients, ",") +
		"\nCLIENT_THREADS=" + strconv.Itoa(traceReplayerConfig.ClientThreads) +
		"\nTRACE_DIRECTORY=" + traceReplayerConfig.TraceDir +
		"\nTRACE_FILES=" + strings.Join(traceReplayerConfig.TraceFiles, ",") +
		"\nWAIT=" + strconv.FormatBool(traceReplayerConfig.Wait) +
		"\nWARMUP_THREADS=" + strconv.Itoa(traceReplayerConfig.WarmupThreads) +
		"\nMASTER_PORT=" + strconv.Itoa(traceReplayerConfig.MasterPort) +
		"\nLIMIT_TYPE=" + traceReplayerConfig.LimitType +
		"\nLIMIT_AMOUNT=" + strconv.Itoa(traceReplayerConfig.LimitAmount) +
		"\nRESULT_DIRECTORY=" + traceReplayerConfig.ResultsDir +
		"\n")
	err := ioutil.WriteFile(traceReplayerConfig.TracePath+"/.env", envVariables, 0644)
	if err != nil {
		return err
	}
	return nil
}
