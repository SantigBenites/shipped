package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ps lists every container whose cgroup still has at least one live process.
func ps() {
	cgRoot := "/sys/fs/cgroup/gocker"
	entries, err := os.ReadDir(cgRoot)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("no containers running")
			return
		}
		must(err)
	}

	fmt.Printf("%-12s  %-8s  %-20s  %s\n", "CONTAINER ID", "PID", "CREATED", "COMMAND")

	running := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()

		// A non-empty cgroup.procs means the container is still alive.
		procs, err := os.ReadFile(filepath.Join(cgRoot, id, "cgroup.procs"))
		if err != nil || len(strings.TrimSpace(string(procs))) == 0 {
			continue
		}

		metaData, err := os.ReadFile(filepath.Join(RUN_DIR, id, "meta"))
		if err != nil {
			continue
		}
		lines := strings.SplitN(strings.TrimSpace(string(metaData)), "\n", 3)
		if len(lines) < 3 {
			continue
		}

		pid := lines[0]
		ts, _ := strconv.ParseInt(lines[1], 10, 64)
		created := time.Unix(ts, 0).Format("2006-01-02 15:04:05")
		command := lines[2]

		fmt.Printf("%-12s  %-8s  %-20s  %s\n", id, pid, created, command)
		running++
	}

	if running == 0 {
		fmt.Println("no containers running")
	}
}
