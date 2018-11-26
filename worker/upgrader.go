package worker

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const infoURL = "http://hash.iptokenmain.com/upgrade/iphash-%s-%s.json"
const upgradeFileName = "upgrade.json"

type upgrader struct {
	upgradeInfo upgradeInfo
	finish      chan upgradeInfo
}

/// Download and decompress new package if found new version
func (this *upgrader) upgrade() {
	// find version file of current version, if not exist,get newest version file from server
	var newUpgradeInfo *upgradeInfo
	exist, err := pathExists(upgradeFileName)
	if err != nil {
		log.Printf("[Error] Check local upgrade information file failed: %#v \n", err)
		this.finish <- this.upgradeInfo
		return
	}
	if exist {
		if this.upgradeInfo.Version == "" {
			data, err := ioutil.ReadFile(upgradeFileName)
			if err != nil {
				log.Printf("[Error] Read local upgrade information file failed: %#v \n", err)
				this.finish <- this.upgradeInfo
				return
			}
			err = json.Unmarshal([]byte(data), &this.upgradeInfo)
			if err != nil {
				log.Printf("[Error] Unmarshal local upgrade information file to json failed: %#v \n", err)
				this.finish <- this.upgradeInfo
				return
			}
		}
	}
	newUpgradeInfo, err = getUpgradeInfo()
	if err != nil {
		log.Printf("[Error] Get upgrade information from server failed: %#v \n", err)
		this.finish <- this.upgradeInfo
		return
	}
	if newUpgradeInfo.Version != this.upgradeInfo.Version {
		log.Println("Found new version of iphash package:", newUpgradeInfo.Version)
		//download and decompress package
		err := downloadAndDecompress(newUpgradeInfo)
		if err != nil {
			log.Printf("[Error] Download and decompress new package failed: %#v \n", err)
			this.finish <- this.upgradeInfo
			return
		}
		//save new upgrade file to disk
		data, err := json.MarshalIndent(newUpgradeInfo, "", "      ")
		if err != nil {
			log.Printf("[Error] Marshal new upgrade information failed: %#v \n", err)
			this.finish <- this.upgradeInfo
			return
		}
		err = ioutil.WriteFile(upgradeFileName, data, 0666)
		if err != nil {
			log.Printf("[Error] Save new upgrade information to disk failed: %#v \n", err)
			this.finish <- this.upgradeInfo
			return
		}
	}
	this.finish <- *newUpgradeInfo
}

/// Check if file exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/// Get upgrade information
func getUpgradeInfo() (*upgradeInfo, error) {
	url := fmt.Sprintf(infoURL, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Get upgrade information failed: %s", resp.Status)
	}
	var result upgradeInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func downloadAndDecompress(upgradeInfo *upgradeInfo) error {
	packageName := fmt.Sprintf("iphash-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, upgradeInfo.Version)
	ret, err := pathExists(packageName)
	if err != nil {
		return err
	}
	needDownload := true
	if ret {
		sha1Digest, err := sha1File(packageName)
		if err != nil {
			return err
		}
		if sha1Digest == upgradeInfo.SHA1 {
			needDownload = false
		} else {
			log.Println("SHA1 digest differ from upgrade information, delete package", packageName)
			err = os.Remove(packageName)
			if err != nil {
				return err
			}
		}
	}
	if needDownload { //download new package
		log.Println("Downloading new package", packageName, "...")
		f, err := os.Create(packageName)
		if err != nil {
			return err
		}
		defer f.Close()
		res, err := http.Get(upgradeInfo.URL)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		_, err = io.Copy(f, res.Body)
		if err != nil {
			return err
		}
		log.Println("New package", packageName, "has been downloaded")
	}

	folderName := fmt.Sprintf("iphash-%s-%s-%s", runtime.GOOS, runtime.GOARCH, upgradeInfo.Version)
	ret, err = pathExists(folderName)
	if err != nil {
		return err
	}
	needDecompress := true
	if ret {
		ret1, _ := pathExists(folderName + "/ipfs-monitor")
		ret2, _ := pathExists(folderName + "/ipfs")
		ret3, _ := pathExists(folderName + "/install.sh")
		if ret1 && ret2 && ret3 {
			needDecompress = false
		} else {
			log.Println("Decompressed files are not completed, remove decompressed folder", folderName)
			err = os.RemoveAll(folderName)
			if err != nil {
				return err
			}
		}
	}
	if needDecompress { // Decompress package
		log.Println("Decompressing package", packageName)
		err = deCompress(packageName, "./")
		if err != nil {
			return err
		}
		os.Chmod(folderName+"/ipfs-monitor", 0755)
		os.Chmod(folderName+"/ipfs", 0755)
		os.Chmod(folderName+"/install.sh", 0755)
		log.Println("Package", packageName, "has been decompressed")
	}
	return nil
}

func sha1File(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func deCompress(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		if strings.HasSuffix(hdr.Name, "/") {
			continue
		}
		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		io.Copy(file, tr)
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}
