package entry

import (
	"iphash-daemon/worker"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/takama/daemon"
)

const (
	name        = "iphash-daemon"
	description = "IPHASH daemon & package update process"
	logFileName = "iphash-daemon.log"
)

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

// Service is the daemon service struct
type Service struct {
	daemon.Daemon
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {
	usage := "Usage: iphash_daemon install | remove | start | stop | status"
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}
	log.Println("-------------------------")
	log.Println("- iphash-daemon started -")
	log.Println("-------------------------")
	setupLog(logFileName)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	executor := &worker.Main{Done: done, Stop: stop}
	go executor.Start()
	killSignal := <-interrupt
	log.Println("Got signal:", killSignal)
	log.Println("iphash-daemon terminating...")
	stop <- struct{}{}
	if killSignal == os.Interrupt {
		<-done
	}
	log.Println("iphash-daemon terminated")
	return "Service exited", nil
}

func Start() {
	srv, err := daemon.New(name, description)
	if err != nil {
		log.Println("[Error]:", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		log.Println(status, "\nError:", err)
		os.Exit(1)
	}

	log.Println(status)
}
