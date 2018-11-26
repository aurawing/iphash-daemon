package worker

import (
	"fmt"
	"log"
	"os"
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
	time.Sleep(time.Second * 10)
	go this.executeMonitor()
	time.Sleep(time.Second * 3)
}

func (this *procManager) init() {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	procInit, err := os.StartProcess(folderName+"/ipfs", []string{"ipfs", "init"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if err == nil {
		procInit.Wait()
	} else {
		log.Println("[Error] IPFS init failed:", err)
	}
}

func (this *procManager) prepare() {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	procPre, err := os.StartProcess(folderName+"/install.sh", []string{"install.sh"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if err == nil {
		procPre.Wait()
	} else {
		log.Println("[Error] install dependencies failed:", err)
	}
}

func (this *procManager) executeIpfs() {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	for !this.stopping {
		procIpfs, err := os.StartProcess(folderName+"/ipfs", []string{"ipfs", "daemon"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
		if err == nil {
			this.ipfs = procIpfs
			procIpfs.Wait()
		} else {
			log.Println("[Error] Error when starting IPFS:", err)
		}
	}
	this.ipfsSig <- struct{}{}
}

func (this *procManager) executeMonitor() {
	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, this.upgradeInfo.Version)
	for !this.stopping {
		procMonitor, err := os.StartProcess(folderName+"/ipfs-monitor", []string{"ipfs-monitor"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
		if err == nil {
			this.monitor = procMonitor
			procMonitor.Wait()
		} else {
			log.Println("[Error] Error when starting IPFS-monitor:", err)
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
