package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Where the container's root filesystem lives on the host.
const ROOTFS = "/tmp/rootfs"

// Runtime state for each container: metadata file lives here.
const RUN_DIR = "/tmp/gocker"

func newID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

// writeMeta persists just enough info for `ps` to display a useful row.
// Format (one value per line): host-pid, unix-timestamp, command string.
func writeMeta(id string, pid int, args []string) {
	dir := filepath.Join(RUN_DIR, id)
	must(os.MkdirAll(dir, 0755))
	content := fmt.Sprintf("%d\n%d\n%s\n", pid, time.Now().Unix(), strings.Join(args, " "))
	must(os.WriteFile(filepath.Join(dir, "meta"), []byte(content), 0644))
}
