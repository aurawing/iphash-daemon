package entry

import (
	"flag"
	"iphash-daemon/worker"
	"log"
	"os"
	"syscall"

	"github.com/marcsauter/single"
	"github.com/sevlyar/go-daemon"
)

const (
	logFileName = "iphash-daemon.log"
	pidFileName = "iphash-daemon.pid"
)

var (
	signal = flag.String("s", "", `send signal to the daemon
		quit - graceful shutdown
		stop - fast shutdown
		reload - reloading the configuration file`)
)

func Start() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	cntxt := &daemon.Context{
		PidFileName: pidFileName,
		PidFilePerm: 0644,
		LogFileName: logFileName,
		LogFilePerm: 0640,
		WorkDir:     "." + string(os.PathSeparator),
		Umask:       027,
		Args:        []string{"[iphash-daemon]"},
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalln("Unable send signal to the daemon:", err)
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	s := single.New("iphash-daemon")
	if err := s.CheckLock(); err != nil && err == single.ErrAlreadyRunning {
		log.Fatal("Another instance of iphash-daemon is already running, exiting")
	} else if err != nil {
		// Another error occurred, might be worth handling it as well
		log.Fatalf("Failed to acquire exclusive app lock: %#v", err)
	}
	defer s.TryUnlock()

	log.Println("-------------------------")
	log.Println("- iphash-daemon started -")
	log.Println("-------------------------")

	setupLog(logFileName)
	//go worker()
	executor := &worker.Main{Done: done, Stop: stop}
	go executor.Start()

	err = daemon.ServeSignals()
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("iphash-daemon terminated")
}

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

// func worker() {
// LOOP:
// 	for {
// 		time.Sleep(time.Second) // this is work to be done by worker.
// 		select {
// 		case <-stop:
// 			break LOOP
// 		default:
// 		}
// 	}
// 	done <- struct{}{}
// }

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	return nil
}
