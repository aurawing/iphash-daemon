package worker

import (
	"bytes"
	"fmt"
	"iphash-daemon/arch"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type procManager struct {
	upgradeInfo upgradeInfo
	monitor     *os.Process
	ipfs        *os.Process
	stopping    bool
	monitorSig  chan struct{}
	ipfsSig     chan struct{}
}

func (this *procManager) boot() {
	this.init()
	this.prepare()
	go this.executeIpfs()
	err := this.check()
	if err != nil {
		log.Printf("[Error] IPFS started failed: %#v \n", err)
	}
	go this.executeMonitor()
	time.Sleep(time.Second * 3)
}

func (this *procManager) init() {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	procInit, err := os.StartProcess(folderName+string(os.PathSeparator)+"ipfs"+arch.ExtExecution(), []string{"ipfs" + arch.ExtExecution(), "init"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if err == nil {
		procInit.Wait()
	} else {
		log.Printf("[Error] IPFS init failed: %#v \n", err)
	}
}

func (this *procManager) prepare() {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	procPre, err := os.StartProcess(folderName+string(os.PathSeparator)+"install"+arch.ExtScript(), []string{"install" + arch.ExtScript()}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if err == nil {
		procPre.Wait()
	} else {
		log.Printf("[Error] install dependencies failed: %#v \n", err)
	}
}

func (this *procManager) check() error {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	i := 30
	for i > 0 {
		cmd := exec.Command(folderName+string(os.PathSeparator)+"ipfs"+arch.ExtExecution(), "stats", "bw")
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err := cmd.Run()
		if err == nil {
			return nil
		} else {
			i = i - 1
		}
		// procCheck, err := os.StartProcess(folderName+"/ipfs", []string{"ipfs", "stats", "bw"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
		// if err == nil {
		// 	procCheck.Wait()
		// 	return nil
		// } else {
		// 	i = i - 1
		// }
		time.Sleep(time.Second * 1)
	}
	return fmt.Errorf("ipfs instance may not properly started")
}

func (this *procManager) executeIpfs() {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	for !this.stopping {
		procIpfs, err := os.StartProcess(folderName+string(os.PathSeparator)+"ipfs"+arch.ExtExecution(), []string{"ipfs" + arch.ExtExecution(), "daemon"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
		if err == nil {
			this.ipfs = procIpfs
			procIpfs.Wait()
		} else {
			log.Printf("[Error] Error when starting IPFS: %#v \n", err)
		}
	}
	this.ipfsSig <- struct{}{}
}

func (this *procManager) executeMonitor() {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	for !this.stopping {
		procMonitor, err := os.StartProcess(folderName+string(os.PathSeparator)+"ipfs-monitor"+arch.ExtExecution(), []string{"ipfs-monitor" + arch.ExtExecution()}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
		if err == nil {
			this.monitor = procMonitor
			procMonitor.Wait()
		} else {
			log.Println("[Error] Error when starting ipfs-monitor: %#v \n", err)
		}
	}
	this.monitorSig <- struct{}{}
}

func (this *procManager) stop() {
	this.stopping = true
	if this.monitor != nil {
		this.monitor.Signal(os.Interrupt)
	}
	if this.ipfs != nil {
		this.ipfs.Signal(os.Interrupt)
	}
LOOP1:
	for {
		select {
		case <-this.monitorSig:
			break LOOP1
		case <-time.After(time.Second * 3):
			this.monitor.Signal(os.Kill)
		}
	}
LOOP2:
	for {
		select {
		case <-this.ipfsSig:
			break LOOP2
		case <-time.After(time.Second * 3):
			this.ipfs.Signal(os.Kill)
		}
	}
}

func (this *procManager) kill() {
	this.stopping = true
	if this.monitor != nil {
		this.monitor.Signal(os.Kill)
	}
	if this.ipfs != nil {
		this.ipfs.Signal(os.Kill)
	}
}
