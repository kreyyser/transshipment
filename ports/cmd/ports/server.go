package main

import (
	"fmt"
	"github.com/kreyyser/transshipment/common/config"
	r "github.com/oklog/run"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server represents ports service structure to handle whole microservice logic
type Server struct {
	ConfigManager *config.ConfigManager
	GRPCServer    *grpc.Server
	DB            *gorm.DB
	worker        r.Group
	lis           net.Listener
}

// NewServer creates a new Server reference
func NewServer(cfmgr *config.ConfigManager) *Server {
	manager := Server{
		ConfigManager: cfmgr,
	}

	manager.initialize()

	return &manager
}

// initialize is responsible of initializing the Server state
func (pt *Server) initialize() {
	// TODO a place for adding any middlewares and GRPC tunings
	pt.GRPCServer = grpc.NewServer()

	reflection.Register(pt.GRPCServer)

	// Setup signal handler for termination
	sigterm := make(chan os.Signal, 1)
	pt.worker.Add(func() error {
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		if _, ok := <-sigterm; ok {
			fmt.Println("msg", "stopping tsst server")
		}
		pt.Shutdown()

		return nil
	}, func(err error) {
		signal.Stop(sigterm)
		pt.Shutdown()
		close(sigterm)
	})
}

// LoadStores loads datastore
func (pt *Server) LoadStores() error {
	// A PLACE TO ADD OTHER DATA STORES
	var db *gorm.DB
	for i := 0; i < 3; i++ {
		var dberr error
		db, dberr = pt.ConfigManager.OpenDB()
		if dberr != nil && i == 2{
			return fmt.Errorf("failed to create postgresql connection: %s", dberr)
		}
		if dberr == nil {
			break
		}

		fmt.Println("giving db time to catch up")
		time.Sleep(time.Second * 2)
	}

	pt.DB = db

	return nil
}

// LoadServices loads all the services under the ports service
func (pt *Server) LoadServices() error {
	conf, err := pt.ConfigManager.GetServiceConfig("ports")
	if err != nil {
		return err
	}

	for _, name := range conf.Services {
		if err := StartService(pt, *conf, name); err != nil {
			fmt.Printf("error loading \"%s\" service", name)
			fmt.Printf("err: %v", err)

			return fmt.Errorf("failed to start service \"%s\"", name)
		}

		fmt.Printf("loaded \"%s\" service", name)
	}

	return nil
}

// Run spins up request handlers
func (pt *Server) Run() error {
	cfg, err := pt.ConfigManager.GetServiceConfig("ports")
	if err != nil {
		fmt.Println(err)
		return err
	}
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		fmt.Printf("failed to start grpc listener. err: %v", err)
	}
	pt.lis = lis

	pt.worker.Add(func() error {
		return pt.GRPCServer.Serve(pt.lis)
	}, func(error) {
		fmt.Println("shutting down grpc server...")
	})

	if err := pt.worker.Run(); err != nil {
		fmt.Println(fmt.Errorf("failure running transport worker: %s", err))
		return err
	}

	return nil
}

// Shutdown is responsible closing services appropriately
func (pt *Server) Shutdown() {
	fmt.Println("shutting down grpc server...")
	pt.GRPCServer.Stop()

	fmt.Println("shutting down db connections")
	if db, err := pt.DB.DB(); err != nil {
		fmt.Println(fmt.Sprintf("failed to close db connection. err: %v", err))
	} else {
		fmt.Println("closing db connections")
		if err := db.Close(); err != nil {
			fmt.Println(fmt.Sprintf("failed to close db connection. err: %v", err))
		}
	}
}