package install

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/wonderivan/logger"
)

//SendPackage is
func (s *SealosInstaller) SendPackage(url string, localPkg bool) {
	pkg := path.Base(url)
	//only http
	isHttp := strings.HasPrefix(url, "http")
	wgetCommand := ""
	if isHttp {
		wgetParam := ""
		if strings.HasPrefix(url, "https") {
			wgetParam = "--no-check-certificate"
		}
		wgetCommand = fmt.Sprintf(" wget %s ", wgetParam)
	}
	remoteCmd := fmt.Sprintf("cd /root &&  %s %s && tar zxvf %s", wgetCommand, url, pkg)
	localCmd := fmt.Sprintf("cd /root && tar zxvf %s ", pkg)
	kubeCmd := "cd /root/kube/shell && sh init.sh"
	kubeLocal := fmt.Sprintf("/root/%s", pkg)

	var wm sync.WaitGroup
	var wn sync.WaitGroup
	for _, master := range s.Masters {
		wm.Add(1)
		go func(master string) {
			defer wm.Done()
			logger.Debug("please wait for tar zxvf exec")
			if isHttp {
				go WatchFileSize(master, kubeLocal)
				Cmd(master, remoteCmd)
			} else {
				if localPkg {
					Copy(master, url, kubeLocal)
				}
				Cmd(master, localCmd)
			}
			Cmd(master, kubeCmd)
		}(master)
	}
	for _, node := range s.Nodes {
		wn.Add(1)
		go func(node string) {
			defer wn.Done()
			logger.Debug("please wait for tar zxvf exec")
			if isHttp {
				go WatchFileSize(node, kubeLocal)
				Cmd(node, remoteCmd)
			} else {
				if localPkg {
					Copy(node, url, kubeLocal)
				}
				Cmd(node, localCmd)
			}
			Cmd(node, kubeCmd)
		}(node)
	}

	wm.Wait()
	wn.Wait()
}
