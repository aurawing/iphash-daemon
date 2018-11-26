package main

import (
	"flag"
	"iphash-daemon/worker"
	"log"
	"os"
	"syscall"

	"github.com/marcsauter/single"
	"github.com/sevlyar/go-daemon"
)

var (
	signal = flag.String("s", "", `send signal to the daemon
		quit - graceful shutdown
		stop - fast shutdown
		reload - reloading the configuration file`)
)

func main() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	cntxt := &daemon.Context{
		PidFileName: "iphash-daemon.pid",
		PidFilePerm: 0644,
		LogFileName: "iphash-daemon.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
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

func main1() {
	// path, err := exec.LookPath("ping")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// proc, err := os.StartProcess(path, []string{"ping", "www.baidu.com"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// go (func() {
	// 	time.Sleep(time.Duration(1) * time.Second)
	// 	proc.Signal(os.Interrupt)
	// })()
	// proc.Wait()
	// fmt.Println("finish")
	// err := worker.DeCompress("test.tar.gz", "/tmp/")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// os.StartProcess("/bin/ping", []string{"ping", "-c 3", "www.baidu.com"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	var (
		stop = make(chan struct{})
		done = make(chan struct{})
	)
	executor := &worker.Main{Done: done, Stop: stop}
	executor.Start()
}
