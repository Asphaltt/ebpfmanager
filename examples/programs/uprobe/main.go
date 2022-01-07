package main

import (
	"github.com/sirupsen/logrus"

	"github.com/ehids/ebpfmanager"
)

var m = &manager.Manager{
	Probes: []*manager.Probe{
		&manager.Probe{
			Section:        "uprobe/readline",
			KernelFuncName: "uprobe_readline",
			BinaryPath:     "/usr/bin/bash",
		},
	},
}

func main() {
	// Initialize the manager
	if err := m.Init(recoverAssets()); err != nil {
		logrus.Fatal(err)
	}

	// Start the manager
	if err := m.Start(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Println("successfully started, head over to /sys/kernel/debug/tracing/trace_pipe")

	// Spawn a bash and right a command to trigger the probe
	if err := trigger(); err != nil {
		logrus.Error(err)
	}

	// Close the manager
	if err := m.Stop(manager.CleanAll); err != nil {
		logrus.Fatal(err)
	}
}
