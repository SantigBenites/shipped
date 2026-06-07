// gocker — a minimal Docker-like container runtime.
//
// It demonstrates the three pillars that make a "container":
//   1. Namespaces       -> isolation (what the process can SEE)
//   2. cgroups          -> limits     (what the process can USE)
//   3. chroot/pivot_root-> filesystem (its own root directory)
//
// MUST be run on Linux, as root. It will NOT work on macOS/Windows
// or inside an unprivileged sandbox — these are real kernel features.
//
// Usage:
//   sudo ./gocker run /bin/sh
//   sudo ./gocker run /bin/echo hello from inside
//   sudo ./gocker ps
//
// You also need a root filesystem to chroot into. The easiest way:
//   mkdir -p /tmp/rootfs
//   cd /tmp/rootfs
//   curl -fsSL https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.0-x86_64.tar.gz | tar -xz
// Then point ROOTFS below at /tmp/rootfs.

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gocker <run|ps> [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("usage: gocker run <command> [args...]")
			os.Exit(1)
		}
		run()
	case "child":
		// Internal: re-exec entry point, already inside new namespaces.
		child()
	case "ps":
		ps()
	default:
		fmt.Printf("unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
