package routes

import (
	"net/http"

	"github.com/txix-open/etp/v3"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"isp-config-service/controller"
)

type Controllers struct {
	Module controller.Module
}

func EndpointDescriptors() []cluster.EndpointDescriptor {
	return endpointDescriptors(Controllers{})
}

func GrpcHandler(wrapper endpoint.Wrapper, c Controllers) *grpc.Mux {
	muxer := grpc.NewMux()
	for _, descriptor := range endpointDescriptors(c) {
		muxer.Handle(descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	return muxer
}

func BindEtp(etpSrv *etp.Server, c Controllers) {
	etpSrv.OnConnect(c.Module.OnConnect)
	etpSrv.OnDisconnect(c.Module.OnDisconnect)
	etpSrv.OnError(c.Module.OnError)
	etpSrv.On(cluster.ModuleSendConfigSchema, etp.HandlerFunc(c.Module.OnModuleConfigSchema))
	etpSrv.On(cluster.ModuleSendRequirements, etp.HandlerFunc(c.Module.OnModuleRequirements))
	etpSrv.On(cluster.ModuleReady, etp.HandlerFunc(c.Module.OnModuleReady))
}

func HttpHandler(etpSrv *etp.Server) http.Handler {
	httpMux := http.NewServeMux()
	httpMux.Handle("/isp-etp/", etpSrv)
	return httpMux
}

func endpointDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{}
}
