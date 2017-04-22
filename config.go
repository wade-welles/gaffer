package main

import (
	"encoding/json"
)

const ZKClass string = "org.apache.zookeeper.server.quorum.QuorumPeerMain"

var (
	// TODO infer all of this God forsaken shit
	DefaultZKClassPath = []string{
		"/usr/share/zookeeper/lib/jline-0.9.94.jar",
		"/usr/share/zookeeper/lilog4j-1.2.16.jar",
		"/usr/share/zookeeper/lib/netty-3.10.5.Final.jar",
		"/usr/share/zookeeper/lib/slf4j-api-1.6.1.jar",
		"/usr/share/zookeeper/lib/slf4j-log4j12-1.6.1.jar",
		"/usr/share/zookeeper/lib/zookeeper-3.4.9.jar",
	}
)

// Opetion represents a process configuration
type Option struct {
	Name  string          `json:"name"`
	Value string          `json:"value"`
	Data  json.RawMessage `json:"data"`
}

// Cluster represents the overall configuration of a Mesos cluster
type Cluster struct {
	Name             string       `json:"name"`
	Initialized      bool         `json:"initialized"`
	ZookeeperReady   bool         `json:"zookeeper_ready"`
	MesosReady       bool         `json:"mesos_ready"`
	Size             int          `json:"size"`
	ZookeeperOptions []*Option    `json:"zookeeper_options"`
	MasterOptions    []*Option    `json:"master_options"`
	AgentOptions     []*Option    `json:"agent_options"`
	Masters          []*Master    `json:"masters"`
	Zookeepers       []*Zookeeper `json:"zookeepers"`
}

// Master represents a single Mesos Master process
type Master struct {
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	Running  bool      `json:"running"`
	Options  []*Option `json:"options"`
}

// Agent represents a single Mesos Agent process
type Agent struct {
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	Running  bool      `json:"running"`
	Options  []*Option `json:"options"`
}

// Zookeeper represents a single Zookeeper process
type Zookeeper struct {
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	Running   bool      `json:"running"`
	Java      string    `json:"java"`
	Classpath []string  `json:"classpath"`
	Options   []*Option `json:"options"`
}
