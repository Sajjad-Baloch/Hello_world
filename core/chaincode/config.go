/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/spf13/viper"
)

const (
	defaultExecutionTimeout = 30 * time.Second
	minimumStartupTimeout   = 5 * time.Second
)

type Config struct {
	TotalQueryLimit int
	TLSEnabled      bool
	Keepalive       time.Duration
	ExecuteTimeout  time.Duration
	StartupTimeout  time.Duration
	LogFormat       string
	LogLevel        string
	ShimLogLevel    string
}

func GlobalConfig() *Config {
	c := &Config{}
	c.load()
	return c
}

func (c *Config) load() {
	viper.SetEnvPrefix("CORE")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	c.TLSEnabled = viper.GetBool("peer.tls.enabled")

	c.Keepalive = toSeconds(viper.GetString("chaincode.keepalive"), 0)
	c.ExecuteTimeout = viper.GetDuration("chaincode.executetimeout")
	if c.ExecuteTimeout < time.Second {
		c.ExecuteTimeout = defaultExecutionTimeout
	}
	c.StartupTimeout = viper.GetDuration("chaincode.startuptimeout")
	if c.StartupTimeout < minimumStartupTimeout {
		c.StartupTimeout = minimumStartupTimeout
	}

	c.LogFormat = viper.GetString("chaincode.logging.format")
	c.LogLevel = getLogLevelFromViper("chaincode.logging.level")
	c.ShimLogLevel = getLogLevelFromViper("chaincode.logging.shim")

	c.TotalQueryLimit = 10000 // need a default just in case it's not set
	if viper.IsSet("ledger.state.totalQueryLimit") {
		c.TotalQueryLimit = viper.GetInt("ledger.state.totalQueryLimit")
	}
}

func toSeconds(s string, def int) time.Duration {
	seconds, err := strconv.Atoi(s)
	if err != nil {
		return time.Duration(def) * time.Second
	}

	return time.Duration(seconds) * time.Second
}

// getLogLevelFromViper gets the chaincode container log levels from viper
func getLogLevelFromViper(key string) string {
	levelString := viper.GetString(key)
	if !flogging.IsValidLevel(levelString) {
		chaincodeLogger.Warningf("%s has invalid log level %s. defaulting to %s", key, levelString, flogging.DefaultLevel())
		levelString = flogging.DefaultLevel()
	}

	return flogging.NameToLevel(levelString).String()
}

// DevModeUserRunsChaincode enables chaincode execution in a development
// environment
const DevModeUserRunsChaincode string = "dev"

// IsDevMode returns true if the peer was configured with development-mode
// enabled.
func IsDevMode() bool {
	mode := viper.GetString("chaincode.mode")

	return mode == DevModeUserRunsChaincode
}
