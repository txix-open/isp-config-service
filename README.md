## Contents
## [1. Installing](#installing)
## [2. Using](#using)
## [3. Environment variables](#environment-variables)
 

  
 
## Installing
### 1. Download dependencies
```
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go get -u github.com/golang/protobuf/protoc-gen-go
```
### 2. Generate GRPC, GRPC-gateway, GRPC-gateway-swagger
###### WINDOWS
```cmd
protoc ^
    -I. -I%GOPATH%\src ^
    -I%GOPATH%\src\github.com\grpc-ecosystem\grpc-gateway\third_party\googleapis ^
    --go_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. --swagger_out=logtostderr=true:. ^
    proto\config_service.proto
```
###### LINUX
```bash
protoc -I. \
	-I$GOPATH \
	-I$GOPATH/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--go_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. --swagger_out=logtostderr=true:. \
	proto/config_service.proto
```

## Using
### 1. GRPC
Default GRPC port is: `5001`, might be changed in config file
Proto file is: `proto/config_service.proto`
### 2. REST proxy to GRPC
Default port is: `5000`, might be changed in config file
Swagger work on http://0.0.0.0:5000/swagger
### 3. SocketIO
For connect to this server two variables must be sent in query path:
1. `module_name` - Module name, for example `auth`.
2. `instance_uuid` - UUID - instance identity.

#### Events
##### Event to a client
`CONFIG:SEND_CONFIG_WHEN_CONNECTED` - When a connection is established to a client will send an event.

`CONFIG:SEND_CONFIG_CHANGED` - When config for particular module had been changed this event with new config will send to a client.

`CONFIG:SEND_CONFIG_ON_REQUEST` - the server will send to a client actual config with this event after `CONFIG:REQUEST_CONFIG`

###### The next events work on both sides. First, when new backend publishes new methods at config service there is incoming one of this events:

`ROUTES:SEND_ROUTES_WHEN_CONNECTED` - A backend sends this event when a connection is established with a config server.

`ROUTES:SEND_ROUTES_CHANGED` - Any module can say that something had changed in its methods.

`ROUTES:SEND_ROUTES_ON_REQUEST` - A config service is able to make a request to renew list methods.

###### A config service, in turn, send the same events types in each its connection.

For all of this events we wait for next data structure:
```json
{
    "address": {
        "ip": "10.10.10.10",
        "port": "5001"
    },
    "endpoints": [
        {"path": "/api/user/get"},
        {"path": "/api/user/save"},
        {"path": "/api/user/delete"}
    ]
}
```
##### Event to the server
`CONFIG:REQUEST_CONFIG` - A client can request current config with this event.

## Environment variables
`APP_PROFILE` - Name for config file, default is `config`. Will seek file with name: `config.yml`.

`APP_CONFIG_PATH` - Absolute path where a config file is. If a variable hadn't been specified config will seek into the same directory, when binary file placed.

`LOG_LEVEL` - Log level, the default is `INFO`.

`APP_MODE` - If set to `dev` logger for SQL will be on and if `LOG_LEVEL` hadn't been specified it will set to `DEBUG`.

## Changes
#### 0.2.2 - 2018-10-05
- Add default config for mdm module
