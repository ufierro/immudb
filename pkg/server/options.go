/*
Copyright 2021 CodeNotary, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codenotary/immudb/pkg/stream"

	"github.com/codenotary/immudb/embedded/store"
	"github.com/codenotary/immudb/pkg/auth"
)

const SystemdbName = "systemdb"
const DefaultdbName = "defaultdb"

// Options server options list
type Options struct {
	Dir                 string
	Network             string
	Address             string
	Port                int
	MetricsPort         int
	Config              string
	Pidfile             string
	Logfile             string
	TLSConfig           *tls.Config
	auth                bool
	MaxRecvMsgSize      int
	NoHistograms        bool
	Detached            bool
	MetricsServer       bool
	WebServer           bool
	WebPort             int
	DevMode             bool
	AdminPassword       string `json:"-"`
	systemAdminDbName   string
	defaultDbName       string
	listener            net.Listener
	usingCustomListener bool
	maintenance         bool
	SigningKey          string
	StoreOptions        *store.Options
	StreamChunkSize     int
	TokenExpiryTimeMin  int
}

// DefaultOptions returns default server options
func DefaultOptions() *Options {
	return &Options{
		Dir:                 "./data",
		Network:             "tcp",
		Address:             "0.0.0.0",
		Port:                3322,
		MetricsPort:         9497,
		WebPort:             8080,
		Config:              "configs/immudb.toml",
		Pidfile:             "",
		Logfile:             "",
		TLSConfig:           &tls.Config{},
		auth:                true,
		MaxRecvMsgSize:      1024 * 1024 * 32, // 32Mb
		NoHistograms:        false,
		Detached:            false,
		MetricsServer:       true,
		WebServer:           true,
		DevMode:             false,
		AdminPassword:       auth.SysAdminPassword,
		systemAdminDbName:   SystemdbName,
		defaultDbName:       DefaultdbName,
		usingCustomListener: false,
		maintenance:         false,
		StoreOptions:        DefaultStoreOptions(),
		StreamChunkSize:     stream.DefaultChunkSize,
		TokenExpiryTimeMin:  1440,
	}
}

func DefaultStoreOptions() *store.Options {
	indexOptions := store.DefaultIndexOptions().WithRenewSnapRootAfter(0)
	return store.DefaultOptions().
		WithIndexOptions(indexOptions).
		WithMaxLinearProofLen(0).
		WithMaxConcurrency(10).
		WithMaxValueLen(32 << 20)
}

// WithDir sets dir
func (o *Options) WithDir(dir string) *Options {
	o.Dir = dir
	return o
}

// WithNetwork sets network
func (o *Options) WithNetwork(network string) *Options {
	o.Network = network
	return o
}

// WithAddress sets address
func (o *Options) WithAddress(address string) *Options {
	o.Address = address
	return o
}

// WithPort sets port
func (o *Options) WithPort(port int) *Options {
	if port > 0 {
		o.Port = port
	}
	return o
}

// WithConfig sets config file name
func (o *Options) WithConfig(config string) *Options {
	o.Config = config
	return o
}

// WithPidfile sets pid file
func (o *Options) WithPidfile(pidfile string) *Options {
	o.Pidfile = pidfile
	return o
}

// WithLogfile sets logfile
func (o *Options) WithLogfile(logfile string) *Options {
	o.Logfile = logfile
	return o
}

// WithTLS sets tls config
func (o *Options) WithTLS(tls *tls.Config) *Options {
	o.TLSConfig = tls
	return o
}

// WithAuth sets auth
func (o *Options) WithAuth(authEnabled bool) *Options {
	o.auth = authEnabled
	return o
}

func (o *Options) WithMaxRecvMsgSize(maxRecvMsgSize int) *Options {
	o.MaxRecvMsgSize = maxRecvMsgSize
	return o
}

// GetAuth gets auth
func (o *Options) GetAuth() bool {
	if o.maintenance {
		return false
	}
	return o.auth
}

// WithNoHistograms disables collection of histograms metrics (e.g. query durations)
func (o *Options) WithNoHistograms(noHistograms bool) *Options {
	o.NoHistograms = noHistograms
	return o
}

// WithDetached sets immudb to be run in background
func (o *Options) WithDetached(detached bool) *Options {
	o.Detached = detached
	return o
}

func (o *Options) WithStoreOptions(storeOpts *store.Options) *Options {
	o.StoreOptions = storeOpts
	return o
}

// Bind returns bind address
func (o *Options) Bind() string {
	return o.Address + ":" + strconv.Itoa(o.Port)
}

// MetricsBind return metrics bind address
func (o *Options) MetricsBind() string {
	return o.Address + ":" + strconv.Itoa(o.MetricsPort)
}

// WebBind return bind address for the Web API/console
func (o *Options) WebBind() string {
	return o.Address + ":" + strconv.Itoa(o.WebPort)
}

// The Web proxy needs to know where to find the GRPC server
func (o *Options) GprcConnectAddr() string {
	if o.Address == "" || o.Address == "0.0.0.0" {
		return "127.0.0.1" + ":" + strconv.Itoa(o.Port)
	}
	return o.Address + ":" + strconv.Itoa(o.Port)
}

// String print options
func (o *Options) String() string {
	rightPad := func(k string, v interface{}) string {
		return fmt.Sprintf("%-17s: %v", k, v)
	}
	opts := make([]string, 0, 17)
	opts = append(opts, "================ Config ================")
	opts = append(opts, rightPad("Data dir", o.Dir))
	opts = append(opts, rightPad("Address", fmt.Sprintf("%s:%d", o.Address, o.Port)))
	if o.MetricsServer {
		opts = append(opts, rightPad("Metrics address", fmt.Sprintf("%s:%d/metrics", o.Address, o.MetricsPort)))
	}
	if o.Config != "" {
		opts = append(opts, rightPad("Config file", o.Config))
	}
	if o.Pidfile != "" {
		opts = append(opts, rightPad("PID file", o.Pidfile))
	}
	if o.Logfile != "" {
		opts = append(opts, rightPad("Log file", o.Logfile))
	}
	opts = append(opts, rightPad("Max recv msg size", o.MaxRecvMsgSize))
	opts = append(opts, rightPad("Auth enabled", o.auth))
	opts = append(opts, rightPad("Dev mode", o.DevMode))
	opts = append(opts, rightPad("Default database", o.defaultDbName))
	opts = append(opts, rightPad("Maintenance mode", o.maintenance))
	opts = append(opts, rightPad("Synced mode", o.StoreOptions.Synced))
	opts = append(opts, "----------------------------------------")
	opts = append(opts, "Superadmin default credentials")
	opts = append(opts, rightPad("   Username", auth.SysAdminUsername))
	opts = append(opts, rightPad("   Password", auth.SysAdminPassword))
	opts = append(opts, "========================================")
	return strings.Join(opts, "\n")
}

// WithMetricsServer ...
func (o *Options) WithMetricsServer(metricsServer bool) *Options {
	o.MetricsServer = metricsServer
	return o
}

// WithWebServer ...
func (o *Options) WithWebServer(webServer bool) *Options {
	o.WebServer = webServer
	return o
}

// WithDevMode ...
func (o *Options) WithDevMode(devMode bool) *Options {
	o.DevMode = devMode
	return o
}

// WithAdminPassword ...
func (o *Options) WithAdminPassword(adminPassword string) *Options {
	o.AdminPassword = adminPassword
	return o
}

//GetSystemAdminDbName returns the System database name
func (o *Options) GetSystemAdminDbName() string {
	return o.systemAdminDbName
}

//GetDefaultDbName returns the default database name
func (o *Options) GetDefaultDbName() string {
	return o.defaultDbName
}

// WithListener used usually to pass a bufered listener for testing purposes
func (o *Options) WithListener(lis net.Listener) *Options {
	o.listener = lis
	o.usingCustomListener = true
	return o
}

// WithMaintenance sets maintenance mode
func (o *Options) WithMaintenance(m bool) *Options {
	o.maintenance = m
	return o
}

// GetMaintenance gets maintenance mode
func (o *Options) GetMaintenance() bool {
	return o.maintenance
}

// WithSigningKey sets signature private key
func (o *Options) WithSigningKey(signingKey string) *Options {
	o.SigningKey = signingKey
	return o
}

// WithStreamChunkSize set the chunk size
func (o *Options) WithStreamChunkSize(streamChunkSize int) *Options {
	o.StreamChunkSize = streamChunkSize
	return o
}

// WithTokenExpiryTime set authentication token expiration time in minutes
func (o *Options) WithTokenExpiryTime(tokenExpiryTimeMin int) *Options {
	o.TokenExpiryTimeMin = tokenExpiryTimeMin
	return o
}
