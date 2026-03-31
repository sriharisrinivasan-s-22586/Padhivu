package model

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/88250/gulu"
	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/util"
)

type LANConnectedDevice struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Host      string `json:"host"`
	OS        string `json:"os"`
	Role      string `json:"role"`
	Connected bool   `json:"connected"`
	LastSeen  int64  `json:"lastSeen"`
}

var (
	lanDevices     = map[string]*LANConnectedDevice{}
	lanDevicesLock = sync.Mutex{}
)

func ensureLANConf() {
	if nil == Conf.Sync.LAN {
		Conf.Sync.LAN = conf.NewLAN()
	}
	if "" == Conf.Sync.LAN.BasePath {
		Conf.Sync.LAN.BasePath = util.DataDir
	}
	if "" == Conf.Sync.LAN.DeviceID {
		Conf.Sync.LAN.DeviceID = Conf.System.ID
	}
	if "" == Conf.Sync.LAN.DeviceName {
		Conf.Sync.LAN.DeviceName = Conf.System.Name
	}
	if 1 > Conf.Sync.LAN.Timeout {
		Conf.Sync.LAN.Timeout = 15
	}
	if 1 > Conf.Sync.LAN.ConcurrentReqs {
		Conf.Sync.LAN.ConcurrentReqs = 4
	}
}

func SetSyncProviderLAN(lan *conf.LAN) (err error) {
	ensureLANConf()

	lan.Endpoint = strings.TrimSpace(lan.Endpoint)
	lan.Endpoint = util.NormalizeEndpoint(lan.Endpoint)
	lan.BasePath = util.NormalizeLocalPath(util.DataDir)
	lan.AuthToken = strings.TrimSpace(lan.AuthToken)
	lan.DeviceID = strings.TrimSpace(lan.DeviceID)
	lan.DeviceName = strings.TrimSpace(lan.DeviceName)
	if "" == lan.DeviceID {
		lan.DeviceID = Conf.System.ID
	}
	if "" == lan.DeviceName {
		lan.DeviceName = Conf.System.Name
	}

	if lan.Serve {
		absPath, absErr := filepath.Abs(lan.BasePath)
		if nil != absErr {
			return absErr
		}
		if !gulu.File.IsExist(absPath) {
			return fmt.Errorf("LAN sync host path [%s] does not exist", lan.BasePath)
		}
		lan.BasePath = absPath
	}

	if !lan.Serve && "" == lan.Endpoint {
		return fmt.Errorf("LAN sync endpoint is required when host mode is disabled")
	}

	lan.Timeout = util.NormalizeTimeout(lan.Timeout)
	if 1 > lan.ConcurrentReqs {
		lan.ConcurrentReqs = 1
	}
	if 16 < lan.ConcurrentReqs {
		lan.ConcurrentReqs = 16
	}

	Conf.Sync.LAN = lan
	Conf.Save()
	registerLANDevice(&LANConnectedDevice{
		ID:        lan.DeviceID,
		Name:      lan.DeviceName,
		Host:      firstServerAddr(),
		OS:        Conf.System.OS,
		Role:      lanRole(),
		Connected: true,
		LastSeen:  util.CurrentTimeMillis(),
	})
	return
}

func LANHostEnabled() bool {
	return nil != Conf.Sync.LAN && Conf.Sync.LAN.Serve && "" != Conf.Sync.LAN.BasePath
}

func lanRole() string {
	if LANHostEnabled() {
		return "host"
	}
	return "client"
}

func registerLANDevice(device *LANConnectedDevice) {
	if nil == device || "" == device.ID {
		return
	}
	lanDevicesLock.Lock()
	defer lanDevicesLock.Unlock()
	device.LastSeen = util.CurrentTimeMillis()
	lanDevices[device.ID] = device
}

func GetLANConnectedDevices() (ret []*LANConnectedDevice) {
	ensureLANConf()
	now := util.CurrentTimeMillis()
	lanDevicesLock.Lock()
	defer lanDevicesLock.Unlock()

	if "" != Conf.Sync.LAN.DeviceID {
		lanDevices[Conf.Sync.LAN.DeviceID] = &LANConnectedDevice{
			ID:        Conf.Sync.LAN.DeviceID,
			Name:      Conf.Sync.LAN.DeviceName,
			Host:      firstServerAddr(),
			OS:        Conf.System.OS,
			Role:      lanRole(),
			Connected: true,
			LastSeen:  now,
		}
	}

	for _, device := range lanDevices {
		copied := *device
		copied.Connected = now-copied.LastSeen < int64((2*time.Minute)/time.Millisecond)
		ret = append(ret, &copied)
	}
	return
}

func firstServerAddr() string {
	if 0 < len(Conf.ServerAddrs) {
		return Conf.ServerAddrs[0]
	}
	return ""
}

func initLANConf() {
	if nil == Conf.Sync.LAN {
		Conf.Sync.LAN = conf.NewLAN()
	}
	ensureLANConf()
	if LANHostEnabled() {
		logging.LogInfof("LAN sync host enabled at [%s]", Conf.Sync.LAN.BasePath)
	}
}
