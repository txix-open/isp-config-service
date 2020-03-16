### v2.2.0
* migrate to go modules
* add linter, refactor

### v2.1.3
* improved logging

### v2.1.2
* leader ws client ConnectionReadLimit is now configured by WS.WsConnectionReadLimitKB config param and defaults to 4 MB

### v2.1.1
* update to new endpoint descriptor struct with extra fields introduced in isp-lib v2.1.0

### v2.1.0
* fix panic type casting in shapshot
* create empty default config now
* add method `/module/get_by_name` to fetch module object by name

### v2.0.3
* fix data race when applyCommandOnLeader in websocket handler
* update dependencies

### v2.0.2
* fix raft server address announcing (serverId == serverAddress == cluster.outerAddress)

### v2.0.1
* fix nil pointer dereference in repositories
* increase default websocket ConnectionReadLimit to 4 MB, add to configuration

### v2.0.0
* full rewrite to support cluster mode
* change websocket protocol from socketio to etp

### v1.1.0
* update libs
* default config handling
