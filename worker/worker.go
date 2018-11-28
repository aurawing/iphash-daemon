package worker

import (
	"log"
	"time"
)

type upgradeInfo struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA1    string `json:"sha1"`
}

type Main struct {
	Stop chan struct{}
	Done chan struct{}
}

func (this *Main) Start() {
	var versionInfo upgradeInfo
	var pManager *procManager
	stop := false
	interval := time.Second * 1
	for !stop {
		select {
		case <-this.Stop: //graceful stop all processes
			log.Println("Stopping iphash-daemon...")
			stop = true
			pManager.stop()
			this.Done <- struct{}{}
		case <-time.After(interval): //call upgrader to check and download new package of iphash
			interval = time.Minute * 10
			finish := make(chan upgradeInfo)
			upgrader := &upgrader{upgradeInfo: versionInfo, finish: finish}
			go upgrader.upgrade()
			newVersionInfo := <-finish
			if newVersionInfo.Version != versionInfo.Version { //Package has upgraded， stop exist progresses and execute new version package
				if pManager != nil {
					pManager.stop()
				}
				pManager = &procManager{upgradeInfo: newVersionInfo, stopping: false, monitorSig: make(chan struct{}), ipfsSig: make(chan struct{})}
				pManager.boot()
				versionInfo = newVersionInfo
			}
		}
	}
}
