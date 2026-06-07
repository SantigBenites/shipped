package main

import (
	"os"
	"path/filepath"
	"strconv"
)

// setupCgroup uses cgroups v2 (the modern unified hierarchy, mounted at
// /sys/fs/cgroup) to cap memory and process count for this container.
// Each container gets its own sub-cgroup so ps can track them independently.
func setupCgroup(id string) {
	cg := filepath.Join("/sys/fs/cgroup/gocker", id)
	must(os.MkdirAll(cg, 0755))

	// Limit to ~100 MB of memory. cgroup v2 expects raw bytes.
	must(os.WriteFile(filepath.Join(cg, "memory.max"),
		[]byte(strconv.Itoa(100*1024*1024)), 0644))

	// Limit to 20 processes (a simple fork-bomb guard).
	must(os.WriteFile(filepath.Join(cg, "pids.max"), []byte("20"), 0644))

	// Move the current process into the cgroup. Everything it spawns inherits the limits.
	must(os.WriteFile(filepath.Join(cg, "cgroup.procs"),
		[]byte(strconv.Itoa(os.Getpid())), 0644))
}
