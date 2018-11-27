package main

import (
	"bytes"
	"flag"
	"fmt"
	"iphash-daemon/worker"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/marcsauter/single"
	"github.com/robfig/cron"
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

func main() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	cntxt := &daemon.Context{
		PidFileName: pidFileName,
		PidFilePerm: 0644,
		LogFileName: logFileName,
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

	setupLog()
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

func setupLog() {
	lf, err := NewLogFile(logFileName, os.Stderr)
	if err != nil {
		log.Fatal("Unable to create log file: ", err)
	}
	log.SetOutput(lf)

	rotateLogSignal := make(chan struct{})
	c := cron.New()
	c.AddFunc("0 0 0 * * ?", func() {
		rotateLogSignal <- struct{}{}
	})
	c.Start()

	// rotate log every 30 seconds.
	//rotateLogSignal := time.Tick(24 * time.Hour)
	go func() {
		for {
			<-rotateLogSignal
			if err := lf.Rotate(); err != nil {
				log.Fatal("Unable to rotate log: ", err)
			}
		}
	}()
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
	// var (
	// 	stop = make(chan struct{})
	// 	done = make(chan struct{})
	// )
	// executor := &worker.Main{Done: done, Stop: stop}
	// executor.Start()

	// procCheck, err := os.StartProcess("/tmp/iphash/iphash-linux-amd64-v0.01/ipfs", []string{"ipfs", "stats", "bw"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	// if err != nil {
	// 	fmt.Printf("%#v", err)
	// }
	// procCheck.Wait()
	// fmt.Println("finish")

	cmd := exec.Command("/tmp/iphash/iphash-linux-amd64-v0.01/ipfs", "stats", "bw")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("out:", outb.String(), "err:", errb.String())
}
