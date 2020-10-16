package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/kreyyser/transshipment/common/config"
	"github.com/kreyyser/transshipment/common/router"
	"github.com/kreyyser/transshipment/restgateway/internal/ports"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	return initServer(cfg)
}

func loadConfig() (*config.ConfigManager, error)  {
	// Handle command line options
	flagset := pflag.NewFlagSet("restgateway", pflag.ExitOnError)
	configPath := flagset.StringP("config", "c", "", "Path to transhipment config yaml")
	_ = flagset.Parse(os.Args[1:])

	//Transhipment config file should exist in the file system
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		return nil, errors.New("no config file provided")
	}

	// Build the Config Manager
	cfmgr := config.NewManager(*configPath)

	// Load configuration file
	if err := cfmgr.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config file: %s", err)
	}

	return cfmgr, nil
}

func signaledCtx() (context.Context, context.CancelFunc) {
	ctx, cncl := context.WithCancel(context.Background())

	{
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			select {
			case <-ctx.Done():
				log.Println("ctx is done")
			case sig := <-sigs:
				log.Printf("signal received: %v", sig)
				cncl()
			}
		}()
	}

	return ctx, cncl
}

func initServer(c *config.ConfigManager) error {
	pc, err := c.GetServiceConfig("ports")
	if err != nil {
		return err
	}
	rg, err := c.GetServiceConfig("restgateway")
	if err != nil {
		return err
	}

	var conn *grpc.ClientConn
	for i := 0; i < 3; i++ {
		var conErr error
		conn, conErr = grpc.Dial(fmt.Sprintf("%s%s", "ports", pc.Address), grpc.WithInsecure())
		if conErr != nil && i == 2{
			return err
		}
		if conErr == nil {
			break
		}

		fmt.Println("giving grpc client time to catch up")
		time.Sleep(time.Second * 2)
	}

	routes := [][]router.Route{
		// TODO a place to add any other modules routes
		ports.Routes(conn),
	}

	var apiRoutes []router.Route
	for _, r := range routes {
		apiRoutes = append(apiRoutes, r...)
	}

	rtr := chi.NewRouter()

	for _, rt := range apiRoutes {
		rtr.MethodFunc(rt.Method, rt.Path, rt.Handler)
	}

	srvr, err := New(rg.Address, rtr)
	if err != nil {
		return err
	}

	ctx, cncl := signaledCtx()
	defer cncl()

	return runSrvr(ctx, srvr)
}

func runSrvr(ctx context.Context, server *Server) error {
	errs := make(chan error, 1)
	go func(s *Server) {
		defer func() {
			_ = s.Close()
		}()
		err := s.Run(ctx)
		if err != nil {
			err = fmt.Errorf("run server error: %w", err)
		}
		errs <- err
	}(server)

	return <-errs
}