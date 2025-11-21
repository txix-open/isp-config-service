// Command rqlited is the rqlite server.
// nolint
package rqlite

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"isp-config-service/service/rqlite/internal/rarchive"
	"isp-config-service/service/rqlite/internal/rtls"

	"github.com/pkg/errors"
	"github.com/rqlite/rqlite-disco-clients/dns"
	"github.com/rqlite/rqlite-disco-clients/dnssrv"
	"github.com/rqlite/rqlite/v9/auth"
	"github.com/rqlite/rqlite/v9/auto/backup"
	"github.com/rqlite/rqlite/v9/cluster"
	"github.com/rqlite/rqlite/v9/cmd"
	"github.com/rqlite/rqlite/v9/db"
	"github.com/rqlite/rqlite/v9/db/extensions"
	httpd "github.com/rqlite/rqlite/v9/http"
	"github.com/rqlite/rqlite/v9/store"
	"github.com/rqlite/rqlite/v9/tcp"
	"github.com/txix-open/isp-kit/config"
)

const name = `rqlite`
const desc = `rqlite is a lightweight, distributed relational database, which uses SQLite as its
storage engine. It provides an easy-to-use, fault-tolerant store for relational data.

Visit https://www.rqlite.io to learn more.`

func init() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stderr)
	log.SetPrefix(fmt.Sprintf("[%s] ", name))
}

type localConfig struct {
	Rqlite *Config
	Backup Backup
}

func main(ctx context.Context, r *Rqlite) error {
	mainCtx := ctx

	cfg, err := ParseFlags(name, desc, &BuildInfo{
		Version:       cmd.Version,
		Commit:        cmd.Commit,
		Branch:        cmd.Branch,
		SQLiteVersion: db.DBVersion,
	})
	if err != nil {
		log.Fatalf("failed to parse command-line flags: %s", err.Error())
	}
	err = evaluateAdvAddresses(r.cfg)
	if err != nil {
		return errors.WithMessage(err, "check adv ports")
	}
	localCfg := localConfig{Rqlite: cfg}
	err = r.cfg.Read(&localCfg)
	if err != nil {
		return errors.WithMessage(err, "read local config")
	}
	dumpRqliteConfig, _ := json.MarshalIndent(cfg, "", "  ")
	log.Printf("rqlite config:\n %s \n", string(dumpRqliteConfig))
	err = cfg.Validate()
	if err != nil {
		return errors.WithMessage(err, "validate rqlite configuration")
	}

	r.localHttpAddr = cfg.HTTPAddr

	// Configure logging and pump out initial message.
	log.Printf("%s starting, version %s, SQLite %s, commit %s, branch %s, compiler (toolchain) %s, compiler (command) %s",
		name, cmd.Version, db.DBVersion, cmd.Commit, cmd.Branch, runtime.Compiler, cmd.CompilerCommand)
	log.Printf("%s, target architecture is %s, operating system target is %s", runtime.Version(),
		runtime.GOARCH, runtime.GOOS)
	log.Printf("launch command: %s", strings.Join(os.Args, " "))

	// Create internode network mux and configure.
	muxLn, err := net.Listen("tcp", cfg.RaftAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", cfg.RaftAddr, err.Error())
	}
	mux, err := startNodeMux(cfg, muxLn)
	if err != nil {
		log.Fatalf("failed to start node mux: %s", err.Error())
	}

	// Raft internode layer
	raftLn := mux.Listen(cluster.MuxRaftHeader)
	raftDialer, err := cluster.CreateRaftDialer(cfg.NodeX509Cert, cfg.NodeX509Key, cfg.NodeX509CACert,
		cfg.NodeVerifyServerName, cfg.NoNodeVerify)
	if err != nil {
		log.Fatalf("failed to create Raft dialer: %s", err.Error())
	}
	raftTn := tcp.NewLayer(raftLn, raftDialer)

	// Create extension store.
	extensionsStore, err := createExtensionsStore(cfg)
	if err != nil {
		log.Fatalf("failed to create extensions store: %s", err.Error())
	}
	extensionsPaths, err := extensionsStore.List()
	if err != nil {
		log.Fatalf("failed to list extensions: %s", err.Error())
	}

	// Create the store.
	str, err := createStore(cfg, raftTn, extensionsPaths)
	if err != nil {
		log.Fatalf("failed to create store: %s", err.Error())
	}

	r.storePtr.Store(str)

	credStr, err := credentialsStore(r.credentials)
	if err != nil {
		log.Fatalf("failed to create credentials: %s", err.Error())
	}
	// Get any credential store.
	/*credStr, err := credentialStore(cfg)
	if err != nil {
		log.Fatalf("failed to get credential store: %s", err.Error())
	}*/

	// Create cluster service now, so nodes will be able to learn information about each other.
	clstrServ, err := clusterService(cfg, mux.Listen(cluster.MuxClusterHeader), str, str, credStr)
	if err != nil {
		log.Fatalf("failed to create cluster service: %s", err.Error())
	}

	// Create the HTTP service.
	//
	// We want to start the HTTP server as soon as possible, so the node is responsive and external
	// systems can see that it's running. We still have to open the Store though, so the node won't
	// be able to do much until that happens however.
	clstrClient, err := createClusterClient(cfg, clstrServ)
	if err != nil {
		log.Fatalf("failed to create cluster client: %s", err.Error())
	}
	httpServ, err := startHTTPService(cfg, str, clstrClient, credStr)
	if err != nil {
		log.Fatalf("failed to start HTTP server: %s", err.Error())
	}

	// Now, open store. How long this takes does depend on how much data is being stored by rqlite.
	if err := str.Open(); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	// Register remaining status providers.
	if err := httpServ.RegisterStatus("cluster", clstrServ); err != nil {
		log.Fatalf("failed to register cluster status provider: %s", err.Error())
	}
	if err := httpServ.RegisterStatus("network", tcp.NetworkReporter{}); err != nil {
		log.Fatalf("failed to register network status provider: %s", err.Error())
	}
	if err := httpServ.RegisterStatus("mux", mux); err != nil {
		log.Fatalf("failed to register mux status provider: %s", err.Error())
	}
	if err := httpServ.RegisterStatus("extensions", extensionsStore); err != nil {
		log.Fatalf("failed to register extensions status provider: %s", err.Error())
	}

	// Create the cluster!
	nodes, err := str.Nodes()
	if err != nil {
		log.Fatalf("failed to get nodes %s", err.Error())
	}
	if err := createCluster(mainCtx, cfg, len(nodes) > 0, clstrClient, str, httpServ, credStr); err != nil {
		log.Fatalf("clustering failure: %s", err.Error())
	}

	// Tell the user the node is ready for HTTP, giving some advice on how to connect.
	log.Printf("node HTTP API available at %s", cfg.HTTPURL())
	h, p, _ := net.SplitHostPort(cfg.HTTPAdv)
	log.Printf("connect using the command-line tool via 'rqlite -H %s -p %s'", h, p)

	// Start any requested auto-backups
	backupSrv, err := startAutoBackups(mainCtx, &localCfg.Backup, str)
	if err != nil {
		log.Fatalf("failed to start auto-backups: %s", err.Error())
	}
	if backupSrv != nil {
		httpServ.RegisterStatus("auto_backups", backupSrv)
	}

	// Block until done.
	<-mainCtx.Done()

	// Stop the HTTP server and other network access first so clients get notification as soon as
	// possible that the node is going away.
	httpServ.Close()
	clstrServ.Close()

	if cfg.RaftClusterRemoveOnShutdown {
		remover := cluster.NewRemover(clstrClient, 5*time.Second, str)
		remover.SetCredentials(cluster.CredentialsFor(credStr, cfg.JoinAs))
		log.Printf("initiating removal of this node from cluster before shutdown")
		if err := remover.Do(cfg.NodeID, true); err != nil {
			log.Fatalf("failed to remove this node from cluster before shutdown: %s", err.Error())
		}
		log.Printf("removed this node successfully from cluster before shutdown")
	}

	if cfg.RaftStepdownOnShutdown {
		if str.IsLeader() {
			// Don't log a confusing message if (probably) not Leader
			log.Printf("stepping down as Leader before shutdown")
		}
		// Perform a stepdown, ignore any errors.
		str.Stepdown(true, "")
	}
	muxLn.Close()

	if err := str.Close(true); err != nil {
		log.Printf("failed to close store: %s", err.Error())
	}

	log.Println("rqlite server stopped")
	return nil
}

func startAutoBackups(ctx context.Context, cfg *Backup, str *store.Store) (*backup.Uploader, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	cfg.SetParams()
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, errors.Errorf("failed to marshal auto-backup settings: %s", err.Error())
	}

	uCfg, sc, err := backup.NewStorageClient(b)
	if err != nil {
		return nil, errors.Errorf("failed to parse auto-backup file: %s", err.Error())
	}
	provider := store.NewProvider(str, uCfg.Vacuum, !uCfg.NoCompress)
	u := backup.NewUploader(sc, provider, time.Duration(uCfg.Interval))
	u.Start(ctx, str.IsLeader)
	return u, nil
}

func evaluateAdvAddresses(cfg *config.Config) error {
	httpAdv := cfg.Optional().String("rqlite.HTTPAdv", "")
	shouldEvalHttp := strings.Contains(httpAdv, "$HOSTNAME")
	raftAdv := cfg.Optional().String("rqlite.RaftAdv", "")
	shouldEvalRaft := strings.Contains(raftAdv, "$HOSTNAME")
	if !shouldEvalHttp && !shouldEvalRaft {
		return nil
	}

	var (
		host string
	)
	data, err := exec.Command("hostname", "-f").Output()
	if err != nil {
		host, err = os.Hostname()
		if err != nil {
			return errors.WithMessage(err, "rqlite.HTTPAdv is not set, couldn't resolve local hostname")
		}
	} else {
		host = strings.TrimSpace(string(data))
	}

	if shouldEvalHttp {
		cfg.Set("rqlite.HTTPAdv", strings.ReplaceAll(httpAdv, "$HOSTNAME", host))
	}
	if shouldEvalRaft {
		cfg.Set("rqlite.RaftAdv", strings.ReplaceAll(raftAdv, "$HOSTNAME", host))
	}
	return nil
}

func createExtensionsStore(cfg *Config) (*extensions.Store, error) {
	str, err := extensions.NewStore(filepath.Join(cfg.DataPath, "extensions"))
	if err != nil {
		log.Fatalf("failed to create extension store: %s", err.Error())
	}

	if len(cfg.ExtensionPaths) > 0 {
		for _, path := range cfg.ExtensionPaths {
			if isDir(path) {
				if err := str.LoadFromDir(path); err != nil {
					log.Fatalf("failed to load extensions from directory: %s", err.Error())
				}
			} else if rarchive.IsZipFile(path) {
				if err := str.LoadFromZip(path); err != nil {
					log.Fatalf("failed to load extensions from zip file: %s", err.Error())
				}
			} else if rarchive.IsTarGzipFile(path) {
				if err := str.LoadFromTarGzip(path); err != nil {
					log.Fatalf("failed to load extensions from tar.gz file: %s", err.Error())
				}
			} else {
				if err := str.LoadFromFile(path); err != nil {
					log.Fatalf("failed to load extension from file: %s", err.Error())
				}
			}
		}
	}

	return str, nil
}

func createStore(cfg *Config, ln *tcp.Layer, extensions []string) (*store.Store, error) {
	dbConf := store.NewDBConfig()
	dbConf.OnDiskPath = cfg.OnDiskPath
	dbConf.FKConstraints = cfg.FKConstraints
	dbConf.Extensions = extensions

	str := store.New(&store.Config{
		DBConf: dbConf,
		Dir:    cfg.DataPath,
		ID:     cfg.NodeID,
	}, ln)

	// Set optional parameters on store.
	str.RaftLogLevel = cfg.RaftLogLevel
	str.ShutdownOnRemove = cfg.RaftShutdownOnRemove
	str.SnapshotThreshold = cfg.RaftSnapThreshold
	str.SnapshotThresholdWALSize = cfg.RaftSnapThresholdWALSize
	str.SnapshotInterval = cfg.RaftSnapInterval
	str.LeaderLeaseTimeout = cfg.RaftLeaderLeaseTimeout
	str.HeartbeatTimeout = cfg.RaftHeartbeatTimeout
	str.ElectionTimeout = cfg.RaftElectionTimeout
	str.ApplyTimeout = cfg.RaftApplyTimeout
	str.BootstrapExpect = cfg.BootstrapExpect
	str.ReapTimeout = cfg.RaftReapNodeTimeout
	str.ReapReadOnlyTimeout = cfg.RaftReapReadOnlyNodeTimeout
	str.AutoVacInterval = cfg.AutoVacInterval
	str.AutoOptimizeInterval = cfg.AutoOptimizeInterval

	if store.IsNewNode(cfg.DataPath) {
		log.Printf("no preexisting node state detected in %s, node may be bootstrapping", cfg.DataPath)
	} else {
		log.Printf("preexisting node state detected in %s", cfg.DataPath)
	}

	return str, nil
}

func startHTTPService(cfg *Config, str *store.Store, cltr *cluster.Client, credStr *auth.CredentialsStore) (*httpd.Service, error) {
	// Create HTTP server and load authentication information.
	s := httpd.New(cfg.HTTPAddr, str, cltr, credStr)

	s.CACertFile = cfg.HTTPx509CACert
	s.CertFile = cfg.HTTPx509Cert
	s.KeyFile = cfg.HTTPx509Key
	s.ClientVerify = cfg.HTTPVerifyClient
	s.DefaultQueueCap = cfg.WriteQueueCap
	s.DefaultQueueBatchSz = cfg.WriteQueueBatchSz
	s.DefaultQueueTimeout = cfg.WriteQueueTimeout
	s.DefaultQueueTx = cfg.WriteQueueTx
	s.BuildInfo = map[string]interface{}{
		"commit":             cmd.Commit,
		"branch":             cmd.Branch,
		"version":            cmd.Version,
		"compiler_toolchain": runtime.Compiler,
		"compiler_command":   cmd.CompilerCommand,
		"build_time":         cmd.Buildtime,
	}
	s.SetAllowOrigin(cfg.HTTPAllowOrigin)
	return s, s.Start()
}

// startNodeMux starts the TCP mux on the given listener, which should be already
// bound to the relevant interface.
func startNodeMux(cfg *Config, ln net.Listener) (*tcp.Mux, error) {
	var err error
	adv := tcp.NameAddress{
		Address: cfg.RaftAdv,
	}

	var mux *tcp.Mux
	if cfg.NodeX509Cert != "" {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("enabling node-to-node encryption with cert: %s, key: %s",
			cfg.NodeX509Cert, cfg.NodeX509Key))
		if cfg.NodeX509CACert != "" {
			b.WriteString(fmt.Sprintf(", CA cert %s", cfg.NodeX509CACert))
		}
		if cfg.NodeVerifyClient {
			b.WriteString(", mutual TLS enabled")
			mux, err = tcp.NewMutualTLSMux(ln, adv, cfg.NodeX509Cert, cfg.NodeX509Key, cfg.NodeX509CACert)
		} else {
			b.WriteString(", mutual TLS disabled")
			mux, err = tcp.NewTLSMux(ln, adv, cfg.NodeX509Cert, cfg.NodeX509Key)
		}
		log.Println(b.String())
	} else {
		mux, err = tcp.NewMux(ln, adv)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create node-to-node mux")
	}
	go mux.Serve()
	return mux, nil
}

func credentialStore(cfg *Config) (*auth.CredentialsStore, error) {
	if cfg.AuthFile == "" {
		return nil, nil
	}
	return auth.NewCredentialsStoreFromFile(cfg.AuthFile)
}

func clusterService(cfg *Config, ln net.Listener, db cluster.Database, mgr cluster.Manager, credStr *auth.CredentialsStore) (*cluster.Service, error) {
	c := cluster.New(ln, db, mgr, credStr)
	c.SetAPIAddr(cfg.HTTPAdv)
	c.EnableHTTPS(cfg.HTTPx509Cert != "" && cfg.HTTPx509Key != "") // Conditions met for an HTTPS API
	if err := c.Open(); err != nil {
		return nil, err
	}
	return c, nil
}

func createClusterClient(cfg *Config, clstr *cluster.Service) (*cluster.Client, error) {
	var dialerTLSConfig *tls.Config
	var err error
	if cfg.NodeX509Cert != "" || cfg.NodeX509CACert != "" {
		dialerTLSConfig, err = rtls.CreateClientConfig(cfg.NodeX509Cert, cfg.NodeX509Key,
			cfg.NodeX509CACert, cfg.NodeVerifyServerName, cfg.NoNodeVerify)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create TLS config for cluster dialer")
		}
	}
	clstrDialer := tcp.NewDialer(cluster.MuxClusterHeader, dialerTLSConfig)
	clstrClient := cluster.NewClient(clstrDialer, cfg.ClusterConnectTimeout)
	if err := clstrClient.SetLocal(cfg.RaftAdv, clstr); err != nil {
		return nil, errors.WithMessage(err, "failed to set cluster client local parameters")
	}
	return clstrClient, nil
}

func createCluster(ctx context.Context, cfg *Config, hasPeers bool, client *cluster.Client, str *store.Store,
	httpServ *httpd.Service, credStr *auth.CredentialsStore) error {
	joins := cfg.JoinAddresses()
	if err := networkCheckJoinAddrs(joins); err != nil {
		return err
	}
	if joins == nil && cfg.DiscoMode == "" && !hasPeers {
		if cfg.RaftNonVoter {
			return errors.Errorf("cannot create a new non-voting node without joining it to an existing cluster")
		}

		// Brand new node, told to bootstrap itself. So do it.
		log.Println("bootstrapping single new node")
		if err := str.Bootstrap(store.NewServer(str.ID(), cfg.RaftAdv, true)); err != nil {
			return errors.WithMessage(err, "failed to bootstrap single new node")
		}
		return nil
	}

	// Prepare definition of being part of a cluster.
	bootDoneFn := func() bool {
		leader, _ := str.LeaderAddr()
		return leader != ""
	}
	clusterSuf := cluster.VoterSuffrage(!cfg.RaftNonVoter)

	joiner := cluster.NewJoiner(client, cfg.JoinAttempts, cfg.JoinInterval)
	joiner.SetCredentials(cluster.CredentialsFor(credStr, cfg.JoinAs))
	if joins != nil && cfg.BootstrapExpect == 0 {
		// Explicit join operation requested, so do it.
		j, err := joiner.Do(ctx, joins, str.ID(), cfg.RaftAdv, clusterSuf)
		if err != nil {
			return errors.WithMessage(err, "failed to join cluster")
		}
		log.Println("successfully joined cluster at", j)
		return nil
	}

	if joins != nil && cfg.BootstrapExpect > 0 {
		// Bootstrap with explicit join addresses requests.
		bs := cluster.NewBootstrapper(cluster.NewAddressProviderString(joins), client)
		bs.SetCredentials(cluster.CredentialsFor(credStr, cfg.JoinAs))
		return bs.Boot(ctx, str.ID(), cfg.RaftAdv, clusterSuf, bootDoneFn, cfg.BootstrapExpectTimeout)
	}

	if cfg.DiscoMode == "" {
		// No more clustering techniques to try. Node will just sit, probably using
		// existing Raft state.
		return nil
	}

	// DNS-based discovery requested. It's OK to proceed with this even if this node
	// is already part of a cluster. Re-joining and re-notifying other nodes will be
	// ignored when the node is already part of the cluster.
	log.Printf("discovery mode: %s", cfg.DiscoMode)
	switch cfg.DiscoMode {
	case DiscoModeDNS, DiscoModeDNSSRV:
		rc := cfg.DiscoConfigReader()
		defer func() {
			if rc != nil {
				rc.Close()
			}
		}()

		var provider interface {
			cluster.AddressProvider
			httpd.StatusReporter
		}
		if cfg.DiscoMode == DiscoModeDNS {
			dnsCfg, err := dns.NewConfigFromReader(rc)
			if err != nil {
				return errors.WithMessage(err, "error reading DNS configuration")
			}
			provider = dns.NewWithPort(dnsCfg, cfg.RaftPort())

		} else {
			dnssrvCfg, err := dnssrv.NewConfigFromReader(rc)
			if err != nil {
				return errors.WithMessage(err, "error reading DNS configuration")
			}
			provider = dnssrv.New(dnssrvCfg)
		}

		bs := cluster.NewBootstrapper(provider, client)
		bs.SetCredentials(cluster.CredentialsFor(credStr, cfg.JoinAs))
		httpServ.RegisterStatus("disco", provider)
		return bs.Boot(ctx, str.ID(), cfg.RaftAdv, clusterSuf, bootDoneFn, cfg.BootstrapExpectTimeout)
	default:
		return errors.Errorf("invalid disco mode %s", cfg.DiscoMode)
	}
}

func networkCheckJoinAddrs(joinAddrs []string) error {
	if len(joinAddrs) > 0 {
		log.Println("checking that supplied join addresses don't serve HTTP(S)")
		if addr, ok := httpd.AnyServingHTTP(joinAddrs); ok {
			return errors.Errorf("join address %s appears to be serving HTTP when it should be Raft", addr)
		}
	}
	return nil
}
