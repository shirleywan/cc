package meta

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jxwr/cc/topo"
	"launchpad.net/gozk"
)

type AppConfig struct {
	AppName      string   `json:"appname"`
	AutoFailover bool     `json:"autofailover"`
	MasterRegion string   `json:"master_region"`
	Regions      []string `json:"regions"`
}

type ControllerConfig struct {
	Ip       string `json:"ip"`
	HttpPort int    `json:"http_port"`
	WsPort   int    `json:"ws_port"`
	Region   string `json:"region"`
}

type FailoverRecord struct {
	AppName   string       `json:"appname"`
	NodeId    string       `json:"node_id"`
	NodeAddr  string       `json:"node_addr"`
	Timestamp time.Time    `json:"timestamp"`
	Region    string       `json:"region"`
	Tag       string       `json:"tag"`
	Ranges    []topo.Range `json:"ranges"`
}

func (m *Meta) handleAppConfigChanged(watch <-chan zookeeper.Event) {
	for {
		event := <-watch
		if event.Type == zookeeper.EVENT_CHANGED {
			a, w, err := m.FetchAppConfig()
			if err == nil {
				m.appConfig = a
				log.Println("meta: app config changed.", a)
			} else {
				log.Printf("meta: fetch app config failed, %v", err)
			}
			watch = w
		} else {
			log.Printf("meta: unexpected event coming, %v", event)
			break
		}
	}
}

func (m *Meta) FetchAppConfig() (*AppConfig, <-chan zookeeper.Event, error) {
	zconn := m.zconn
	appName := m.appName
	data, _, watch, err := zconn.GetW("/r3/app/" + appName)
	if err != nil {
		return nil, watch, err
	}
	var c AppConfig
	err = json.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, watch, fmt.Errorf("meta: parse app config error, %v", err)
	}
	if c.AppName != appName {
		return nil, watch, fmt.Errorf("meta: local appname is different from zk, %s <-> %s", appName, c.AppName)
	}
	if c.MasterRegion == "" {
		return nil, watch, fmt.Errorf("meta: master region not set")
	}
	if len(c.Regions) == 0 {
		return nil, watch, fmt.Errorf("meta: regions empty")
	}
	return &c, watch, nil
}

func (m *Meta) RegisterLocalController() error {
	zconn := m.zconn
	zkPath := fmt.Sprintf(m.ccDirPath + "/cc_" + m.localRegion + "_")
	conf := &ControllerConfig{
		Ip:       m.localIp,
		HttpPort: m.httpPort,
		Region:   m.localRegion,
		WsPort:   m.wsPort,
	}
	data, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	path, err := zconn.Create(zkPath, string(data), zookeeper.SEQUENCE|zookeeper.EPHEMERAL, zookeeper.WorldACL(PERM_FILE))
	if err == nil {
		xs := strings.Split(path, "/")
		m.selfZNodeName = xs[len(xs)-1]
	}
	return err
}

func (m *Meta) FetchControllerConfig(zkNode string) (*ControllerConfig, <-chan zookeeper.Event, error) {
	data, _, watch, err := m.zconn.GetW(m.ccDirPath + "/" + zkNode)
	if err != nil {
		return nil, watch, err
	}
	var c ControllerConfig
	err = json.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, watch, err
	}
	return &c, watch, nil
}

func (m *Meta) IsDoingFailover() (bool, error) {
	stat, err := m.zconn.Exists("/r3/failover/doing")
	if err == nil {
		if stat != nil {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return true, err
	}
}

func (m *Meta) LastFailoverRecord() (*FailoverRecord, error) {
	children, stat, err := m.zconn.Children("/r3/failover/history")
	if err != nil {
		return nil, err
	}
	if stat.NumChildren() == 0 {
		return nil, nil
	}

	last := children[len(children)-1]
	data, _, err := m.zconn.Get("/r3/failover/history/" + last)
	if err != nil {
		return nil, err
	}

	var record FailoverRecord
	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (m *Meta) AddFailoverRecord(record *FailoverRecord) error {
	zkPath := fmt.Sprintf("/r3/failover/history/record_%s_%s", record.AppName, record.Region)
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	path, err := m.zconn.Create(zkPath, string(data), zookeeper.SEQUENCE, zookeeper.WorldACL(PERM_FILE))
	log.Printf("meta: failover record created at %s", path)
	return nil
}
