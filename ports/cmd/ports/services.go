package main

import (
	"fmt"
	"github.com/kreyyser/transshipment/common/config"
	"github.com/kreyyser/transshipment/ports/internal/services"
	"github.com/kreyyser/transshipment/ports/internal/services/portentries"
)

var initializers = map[string]initializer{
	"portentries": func(ctx *initCtx) interface{} {
		return portentries.NewPortsService(ctx.DB)
	},
}

type initializer func(ctx *initCtx) interface{}

type initCtx struct {
	*Server
	conf config.ServiceConfig
}

// StartService is responsible for loading each ports service into the registry
func StartService(nx *Server, conf config.ServiceConfig, name string) error {
	if init, ok := initializers[name]; ok {
		ctx := &initCtx{
			Server: nx,
			conf:   conf,
		}
		return ctx.register(init(ctx))
	}
	return fmt.Errorf("no service %q found", name)
}

func (ctx *initCtx) register(v interface{}) error {
	gSVC, ok := v.(services.GRPCService)
	if !ok {
		return fmt.Errorf("%T is not GRPCService, but wanted gRPC", v)
	}
	gSVC.GRPCRegister(ctx.GRPCServer)

	return nil
}
