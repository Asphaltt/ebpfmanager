package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"

	"github.com/ehids/ebpfmanager"
)

var m1 = &manager.Manager{
	Probes: []*manager.Probe{
		&manager.Probe{
			UID:     "MyVFSMkdir1",
			Section: "kprobe/vfs_mkdir",
			MatchFuncName: "kprobe_vfs_mkdir",
		},
		&manager.Probe{
			Section: "kprobe/vfs_opennnnnn",
			MatchFuncName: "kprobe_open",
		},
		&manager.Probe{
			Section: "kprobe/exclude",
			MatchFuncName: "kprobe_exclude",
		},
	},
}

var options1 = manager.Options{
	ActivatedProbes: []manager.ProbesSelector{
		&manager.ProbeSelector{
			ProbeIdentificationPair: manager.ProbeIdentificationPair{
				UID:     "MyVFSMkdir1",
				MatchFuncName: "kprobe_vfs_mkdir",
			},
		},
		&manager.AllOf{
			Selectors: []manager.ProbesSelector{
				&manager.ProbeSelector{
					ProbeIdentificationPair: manager.ProbeIdentificationPair{
						UID:     "MyVFSMkdir1",
						MatchFuncName: "kprobe_vfs_mkdir",
					},
				},
			},
		},
		&manager.OneOf{
			Selectors: []manager.ProbesSelector{
				&manager.ProbeSelector{
					ProbeIdentificationPair: manager.ProbeIdentificationPair{
						MatchFuncName: "kprobe_open",
					},
				},
				&manager.ProbeSelector{
					ProbeIdentificationPair: manager.ProbeIdentificationPair{
						MatchFuncName: "kprobe_exclude",
					},
				},
			},
		},
		&manager.BestEffort{
			Selectors: []manager.ProbesSelector{
				&manager.ProbeSelector{
					ProbeIdentificationPair: manager.ProbeIdentificationPair{
						MatchFuncName: "kprobe_open",
					},
				},
				&manager.ProbeSelector{
					ProbeIdentificationPair: manager.ProbeIdentificationPair{
						MatchFuncName: "kprobe_exclude",
					},
				},
			},
		}},
	ExcludedMatchFuns: []string{
		"kprobe_exclude",
	},
}

var m2 = &manager.Manager{
	Probes: []*manager.Probe{
		&manager.Probe{
			UID:     "MyVFSMkdir2",
			MatchFuncName: "kprobe/vfs_mkdir",
			Section: "kprobe_vfs_mkdir",
		},
		&manager.Probe{
			//Section: "kprobe/vfs_opennnnnn",
			MatchFuncName: "kprobe_vfs_open",
		},
		&manager.Probe{
			//Section: "kprobe/exclude",
			MatchFuncName: "kprobe_exclude",
		},
	},
}

var options2 = manager.Options{
	ActivatedProbes: []manager.ProbesSelector{
		&manager.ProbeSelector{
			ProbeIdentificationPair: manager.ProbeIdentificationPair{
				UID:     "MyVFSMkdir2",
				MatchFuncName: "kprobe_vfs_mkdir",
			},
		},
		&manager.AllOf{
			Selectors: []manager.ProbesSelector{
				&manager.ProbeSelector{
					ProbeIdentificationPair: manager.ProbeIdentificationPair{
						MatchFuncName: "kprobe_vfs_open",
					},
				},
			},
		},
		&manager.OneOf{
			Selectors: []manager.ProbesSelector{
				&manager.ProbeSelector{
					ProbeIdentificationPair: manager.ProbeIdentificationPair{
						MatchFuncName: "kprobe_vfs_open",
					},
				},
				&manager.ProbeSelector{
					ProbeIdentificationPair: manager.ProbeIdentificationPair{
						MatchFuncName: "kprobe_exclude",
					},
				},
			},
		},
	},
	ExcludedMatchFuns: []string{
		"kprobe_exclude",
	},
}

var m3 = &manager.Manager{
	Probes: []*manager.Probe{
		&manager.Probe{
			UID:     "MyVFSMkdir2",
			Section: "kprobe/vfs_mkdir",
		},
		&manager.Probe{
			Section: "kprobe/vfs_opennnnnn",
		},
		&manager.Probe{
			Section: "kprobe/exclude",
		},
	},
}

func main() {
	// Initialize the managers
	logrus.Printf("Kprobe/exclude2 start...")
	if err := m1.InitWithOptions(recoverAssets(), options1); err != nil {
		logrus.Fatal(err)
	}

	// Start m1
	logrus.Printf("m1.Start()...")
	if err := m1.Start(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Println("m1 successfully started")

	// Create a folder to trigger the probes
	if err := trigger(); err != nil {
		logrus.Error(err)
	}

	if err := m1.Stop(manager.CleanAll); err != nil {
		logrus.Fatal(err)
	}

	logrus.Println("=> Cmd+C to continue")
	wait()

	logrus.Println("moving on to m2 (an error is expected)")
	// Initialize the managers
	if err := m2.InitWithOptions(recoverAssets(), options2); err != nil {
		logrus.Fatal(err)
	}

	// Start m2
	if err := m2.Start(); err != nil {
		logrus.Error(err)
	}

	logrus.Println("=> Cmd+C to continue")
	wait()

	logrus.Println("moving on to m3 (an error is expected)")
	if err := m3.Init(recoverAssets()); err != nil {
		logrus.Fatal(err)
	}

	// Start m3
	if err := m3.Start(); err != nil {
		logrus.Error(err)
	}

	logrus.Println("updating activated probes of m3 (no error is expected)")

	mkdirID := manager.ProbeIdentificationPair{UID: "MyVFSMkdir2", MatchFuncName: "kprobe_vfs_mkdir"}
	if err := m3.UpdateActivatedProbes([]manager.ProbesSelector{
		&manager.ProbeSelector{
			ProbeIdentificationPair: mkdirID,
		},
	}); err != nil {
		logrus.Error(err)
	}

	vfsOpenID := manager.ProbeIdentificationPair{MatchFuncName: "kprobe_vfs_open"}
	vfsOpenProbe, ok := m3.GetProbe(vfsOpenID)
	if !ok {
		logrus.Fatal("Failed to find kprobe/vfs_opennnnnn")
	}

	if vfsOpenProbe.Enabled {
		logrus.Errorf("kprobe/vfs_opennnnnn should not be enabled")
	}
}

// wait - Waits until an interrupt or kill signal is sent
func wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
	fmt.Println()
}
