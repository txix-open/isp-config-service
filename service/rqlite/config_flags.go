// nolint
package rqlite

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Config represents the configuration as set by command-line flags.
// All variables will be set, unless explicit noted.
type Config struct {
	// DataPath is path to node data. Always set.
	DataPath string

	// ExtensionPaths is a comma-delimited list of path to SQLite extensions to be loaded.
	// Each element may be a directory path, zipfile, or tar.gz file. May not be set.
	ExtensionPaths StringSliceValue

	// HTTPAddr is the bind network address for the HTTP Server.
	// It never includes a trailing HTTP or HTTPS.
	HTTPAddr string

	// HTTPAdv is the advertised HTTP server network.
	HTTPAdv string

	// HTTPAllowOrigin is the value to set for Access-Control-Allow-Origin HTTP header.
	HTTPAllowOrigin string

	// AuthFile is the path to the authentication file. May not be set.
	AuthFile string `filepath:"true"`

	// AutoBackupFile is the path to the auto-backup file. May not be set.
	AutoBackupFile string `filepath:"true"`

	// AutoRestoreFile is the path to the auto-restore file. May not be set.
	AutoRestoreFile string `filepath:"true"`

	// HTTPx509CACert is the path to the CA certificate file for when this node verifies
	// other certificates for any HTTP communications. May not be set.
	HTTPx509CACert string `filepath:"true"`

	// HTTPx509Cert is the path to the X509 cert for the HTTP server. May not be set.
	HTTPx509Cert string `filepath:"true"`

	// HTTPx509Key is the path to the private key for the HTTP server. May not be set.
	HTTPx509Key string `filepath:"true"`

	// HTTPVerifyClient indicates whether the HTTP server should verify client certificates.
	HTTPVerifyClient bool

	// NodeX509CACert is the path to the CA certificate file for when this node verifies
	// other certificates for any inter-node communications. May not be set.
	NodeX509CACert string `filepath:"true"`

	// NodeX509Cert is the path to the X509 cert for the Raft server. May not be set.
	NodeX509Cert string `filepath:"true"`

	// NodeX509Key is the path to the X509 key for the Raft server. May not be set.
	NodeX509Key string `filepath:"true"`

	// NoNodeVerify disables checking other nodes' Node X509 certs for validity.
	NoNodeVerify bool

	// NodeVerifyClient enable mutual TLS for node-to-node communication.
	NodeVerifyClient bool

	// NodeVerifyServerName is the hostname to verify on the certificates returned by nodes.
	// If NoNodeVerify is true this field is ignored.
	NodeVerifyServerName string

	// NodeID is the Raft ID for the node.
	NodeID string

	// RaftAddr is the bind network address for the Raft server.
	RaftAddr string

	// RaftAdv is the advertised Raft server address.
	RaftAdv string

	// JoinAddrs is the list of Raft addresses to use for a join attempt.
	JoinAddrs string

	// JoinAttempts is the number of times a node should attempt to join using a
	// given address.
	JoinAttempts int

	// JoinInterval is the time between retrying failed join operations.
	JoinInterval time.Duration

	// JoinAs sets the user join attempts should be performed as. May not be set.
	JoinAs string

	// BootstrapExpect is the minimum number of nodes required for a bootstrap.
	BootstrapExpect int

	// BootstrapExpectTimeout is the maximum time a bootstrap operation can take.
	BootstrapExpectTimeout time.Duration

	// DiscoMode sets the discovery mode. May not be set.
	DiscoMode string

	// DiscoKey sets the discovery prefix key.
	DiscoKey string

	// DiscoConfig sets the path to any discovery configuration file. May not be set.
	DiscoConfig string

	// OnDiskPath sets the path to the SQLite file. May not be set.
	OnDiskPath string

	// FKConstraints enables SQLite foreign key constraints.
	FKConstraints bool

	// AutoVacInterval sets the automatic VACUUM interval. Use 0s to disable.
	AutoVacInterval time.Duration

	// AutoOptimizeInterval sets the automatic optimization interval. Use 0s to disable.
	AutoOptimizeInterval time.Duration

	// RaftLogLevel sets the minimum logging level for the Raft subsystem.
	RaftLogLevel string

	// RaftNonVoter controls whether this node is a voting, read-only node.
	RaftNonVoter bool

	// RaftSnapThreshold is the number of outstanding log entries that trigger snapshot.
	RaftSnapThreshold uint64

	// RaftSnapThreshold is the size of a SQLite WAL file which will trigger a snapshot.
	RaftSnapThresholdWALSize uint64

	// RaftSnapInterval sets the threshold check interval.
	RaftSnapInterval time.Duration

	// RaftLeaderLeaseTimeout sets the leader lease timeout.
	RaftLeaderLeaseTimeout time.Duration

	// RaftHeartbeatTimeout specifies the time in follower state without contact
	// from a Leader before the node attempts an election.
	RaftHeartbeatTimeout time.Duration

	// RaftElectionTimeout specifies the time in candidate state without contact
	// from a Leader before the node attempts an election.
	RaftElectionTimeout time.Duration

	// RaftApplyTimeout sets the Log-apply timeout.
	RaftApplyTimeout time.Duration

	// RaftShutdownOnRemove sets whether Raft should be shutdown if the node is removed
	RaftShutdownOnRemove bool

	// RaftClusterRemoveOnShutdown sets whether the node should remove itself from the cluster on shutdown
	RaftClusterRemoveOnShutdown bool

	// RaftStepdownOnShutdown sets whether Leadership should be relinquished on shutdown
	RaftStepdownOnShutdown bool

	// RaftReapNodeTimeout sets the duration after which a non-reachable voting node is
	// reaped i.e. removed from the cluster.
	RaftReapNodeTimeout time.Duration

	// RaftReapReadOnlyNodeTimeout sets the duration after which a non-reachable non-voting node is
	// reaped i.e. removed from the cluster.
	RaftReapReadOnlyNodeTimeout time.Duration

	// ClusterConnectTimeout sets the timeout when initially connecting to another node in
	// the cluster, for non-Raft communications.
	ClusterConnectTimeout time.Duration

	// WriteQueueCap is the default capacity of Execute queues
	WriteQueueCap int

	// WriteQueueBatchSz is the default batch size for Execute queues
	WriteQueueBatchSz int

	// WriteQueueTimeout is the default time after which any data will be sent on
	// Execute queues, if a batch size has not been reached.
	WriteQueueTimeout time.Duration

	// WriteQueueTx controls whether writes from the queue are done within a transaction.
	WriteQueueTx bool

	// CPUProfile enables CPU profiling.
	CPUProfile string

	// MemProfile enables memory profiling.
	MemProfile string

	// TraceProfile enables trace profiling.
	TraceProfile string
}

type Backup struct {
	Enabled           bool
	Amount            int
	Version           int
	Type              string
	Interval          string
	ClearIntervalDays int
	Timestamp         bool
	Vacuum            bool
	Sub               Sub
}

type Sub struct {
	Dir  string
	Name string
}

// Validate checks the configuration for internal consistency, and activates
// important rqlite policies. It must be called at least once on a Config
// object before the Config object is used. It is OK to call more than
// once.
func (c *Config) Validate() error {
	dataPath, err := filepath.Abs(c.DataPath)
	if err != nil {
		return errors.Errorf("failed to determine absolute data path: %s", err.Error())
	}
	c.DataPath = dataPath

	err = c.CheckFilePaths()
	if err != nil {
		return err
	}

	err = c.CheckDirPaths()
	if err != nil {
		return err
	}

	if len(c.ExtensionPaths) > 0 {
		for _, p := range c.ExtensionPaths {
			if !fileExists(p) {
				return errors.Errorf("extension path does not exist: %s", p)
			}
		}
	}

	if !bothUnsetSet(c.HTTPx509Cert, c.HTTPx509Key) {
		return errors.Errorf("either both -%s and -%s must be set, or neither", HTTPx509CertFlag, HTTPx509KeyFlag)
	}
	if !bothUnsetSet(c.NodeX509Cert, c.NodeX509Key) {
		return errors.Errorf("either both -%s and -%s must be set, or neither", NodeX509CertFlag, NodeX509KeyFlag)

	}

	if c.RaftAddr == c.HTTPAddr {
		return errors.New("HTTP and Raft addresses must differ")
	}

	// Enforce policies regarding addresses
	if c.RaftAdv == "" {
		c.RaftAdv = c.RaftAddr
	}
	if c.HTTPAdv == "" {
		c.HTTPAdv = c.HTTPAddr
	}

	// Node ID policy
	if c.NodeID == "" {
		c.NodeID = c.RaftAdv
	}

	// Perform some address validity checks.
	if strings.HasPrefix(strings.ToLower(c.HTTPAddr), "http") ||
		strings.HasPrefix(strings.ToLower(c.HTTPAdv), "http") {
		return errors.New("HTTP options should not include protocol (http:// or https://)")
	}
	if _, _, err := net.SplitHostPort(c.HTTPAddr); err != nil {
		return errors.New("HTTP bind address not valid")
	}

	hadv, _, err := net.SplitHostPort(c.HTTPAdv)
	if err != nil {
		return errors.New("HTTP advertised HTTP address not valid")
	}
	if addr := net.ParseIP(hadv); addr != nil && addr.IsUnspecified() {
		return errors.Errorf("advertised HTTP address is not routable (%s), specify it via -%s or -%s",
			hadv, HTTPAddrFlag, HTTPAdvAddrFlag)
	}

	if _, rp, err := net.SplitHostPort(c.RaftAddr); err != nil {
		return errors.New("raft bind address not valid")
	} else if _, err := strconv.Atoi(rp); err != nil {
		return errors.New("raft bind port not valid")
	}

	radv, rp, err := net.SplitHostPort(c.RaftAdv)
	if err != nil {
		return errors.New("raft advertised address not valid")
	}
	if addr := net.ParseIP(radv); addr != nil && addr.IsUnspecified() {
		return errors.Errorf("advertised Raft address is not routable (%s), specify it via -%s or -%s",
			radv, RaftAddrFlag, RaftAdvAddrFlag)
	}
	if _, err := strconv.Atoi(rp); err != nil {
		return errors.New("raft advertised port is not valid")
	}

	if c.RaftAdv == c.HTTPAdv {
		return errors.New("advertised HTTP and Raft addresses must differ")
	}

	// Enforce bootstrapping policies
	if c.BootstrapExpect > 0 && c.RaftNonVoter {
		return errors.New("bootstrapping only applicable to voting nodes")
	}

	// Join parameters OK?
	if c.JoinAddrs != "" {
		addrs := strings.Split(c.JoinAddrs, ",")
		for i := range addrs {
			if _, _, err := net.SplitHostPort(addrs[i]); err != nil {
				return errors.Errorf("%s is an invalid join address", addrs[i])
			}

			if c.BootstrapExpect == 0 {
				if addrs[i] == c.RaftAdv || addrs[i] == c.RaftAddr {
					return errors.New("node cannot join with itself unless bootstrapping")
				}
				if c.AutoRestoreFile != "" {
					return errors.New("auto-restoring cannot be used when joining a cluster")
				}
			}
		}

		if c.DiscoMode != "" {
			return errors.New("disco mode cannot be used when also explicitly joining a cluster")
		}
	}

	// Valid disco mode?
	switch c.DiscoMode {
	case "":
	case DiscoModeEtcdKV, DiscoModeConsulKV:
		if c.BootstrapExpect > 0 {
			return errors.Errorf("bootstrapping not applicable when using %s", c.DiscoMode)
		}
	case DiscoModeDNS, DiscoModeDNSSRV:
		if c.BootstrapExpect == 0 && !c.RaftNonVoter {
			return errors.Errorf("bootstrap-expect value required when using %s with a voting node", c.DiscoMode)
		}
	default:
		return errors.Errorf("disco mode must be one of %s, %s, %s, or %s",
			DiscoModeConsulKV, DiscoModeEtcdKV, DiscoModeDNS, DiscoModeDNSSRV)
	}

	return nil
}

// JoinAddresses returns the join addresses set at the command line. Returns nil
// if no join addresses were set.
func (c *Config) JoinAddresses() []string {
	if c.JoinAddrs == "" {
		return nil
	}
	return strings.Split(c.JoinAddrs, ",")
}

// HTTPURL returns the fully-formed, advertised HTTP API address for this config, including
// protocol, host and port.
func (c *Config) HTTPURL() string {
	apiProto := "http"
	if c.HTTPx509Cert != "" {
		apiProto = "https"
	}
	return fmt.Sprintf("%s://%s", apiProto, c.HTTPAdv)
}

// RaftPort returns the port on which the Raft system is listening. Validate must
// have been called before calling this method.
func (c *Config) RaftPort() int {
	_, port, err := net.SplitHostPort(c.RaftAddr)
	if err != nil {
		panic("RaftAddr not valid")
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		panic("RaftAddr port not valid")
	}
	return p
}

// DiscoConfigReader returns a ReadCloser providing access to the Disco config.
// The caller must call close on the ReadCloser when finished with it. If no
// config was supplied, it returns nil.
func (c *Config) DiscoConfigReader() io.ReadCloser {
	var rc io.ReadCloser
	if c.DiscoConfig == "" {
		return nil
	}

	// Open config file. If opening fails, assume string is the literal config.
	cfgFile, err := os.Open(c.DiscoConfig)
	if err != nil {
		rc = io.NopCloser(bytes.NewReader([]byte(c.DiscoConfig)))
	} else {
		rc = cfgFile
	}
	return rc
}

// CheckFilePaths checks that all file paths in the config exist.
// Empty filepaths are ignored.
func (c *Config) CheckFilePaths() error {
	v := reflect.ValueOf(c).Elem()

	// Iterate through the fields of the struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldValue := v.Field(i)

		if fieldValue.Kind() != reflect.String {
			continue
		}

		if tagValue, ok := field.Tag.Lookup("filepath"); ok && tagValue == "true" {
			filePath := fieldValue.String()
			if filePath == "" {
				continue
			}
			if !fileExists(filePath) {
				return errors.Errorf("%s does not exist", filePath)
			}
		}
	}
	return nil
}

// CheckDirPaths checks that all directory paths in the config exist and are directories.
// Empty directory paths are ignored.
func (c *Config) CheckDirPaths() error {
	v := reflect.ValueOf(c).Elem()

	// Iterate through the fields of the struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldValue := v.Field(i)

		if fieldValue.Kind() != reflect.String {
			continue
		}

		if tagValue, ok := field.Tag.Lookup("dirpath"); ok && tagValue == "true" {
			dirPath := fieldValue.String()
			if dirPath == "" {
				continue
			}
			if !fileExists(dirPath) {
				return errors.Errorf("%s does not exist", dirPath)
			}
			if !isDir(dirPath) {
				return errors.Errorf("%s is not a directory", dirPath)
			}
		}
	}
	return nil
}

// bothUnsetSet returns true if both a and b are unset, or both are set.
func bothUnsetSet(a, b string) bool {
	return (a == "" && b == "") || (a != "" && b != "")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
