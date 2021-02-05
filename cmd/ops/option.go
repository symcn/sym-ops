package ops

import "time"

type controllerOption struct {
	Threadiness int
	MetricPort  int
	PprofPort   int
	Qos         int
	Burst       int

	LeaderElection          bool
	LeaderElectionNamespace string
	LeaderElectionID        string

	SyncPeriod          time.Duration
	HealthCheckInterval time.Duration
	ExecTimeout         time.Duration
}

func defaultCtrlOption() *controllerOption {
	return &controllerOption{
		Threadiness:             1,
		MetricPort:              9090,
		PprofPort:               34901,
		Qos:                     80,
		Burst:                   100,
		LeaderElection:          false,
		LeaderElectionNamespace: "",
		LeaderElectionID:        "",
		SyncPeriod:              time.Minute * 30,
		HealthCheckInterval:     time.Second * 10,
		ExecTimeout:             time.Second * 5,
	}
}
