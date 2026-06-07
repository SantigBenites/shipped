package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// run is the parent. It re-executes this same binary as `gocker child ...`
// but asks the kernel to place that child into brand-new namespaces.
func run() {
	id := newID()
	fmt.Printf("[parent] container %s  pid %d  launching child in new namespaces\n", id, os.Getpid())

	// /proc/self/exe is a symlink to the currently running binary, so this
	// is the "re-exec yourself" trick that works around Go's no-fork limitation.
	// We insert the container ID as the second argument so child() can read it.
	cmd := exec.Command("/proc/self/exe", append([]string{"child", id}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Each CLONE_NEW* flag creates one new namespace for the child:
		Cloneflags: syscall.CLONE_NEWUTS | // hostname
			syscall.CLONE_NEWPID | // process IDs (child sees itself as PID 1)
			syscall.CLONE_NEWNS | // mount points
			syscall.CLONE_NEWNET | // network interfaces (container starts with none)
			syscall.CLONE_NEWIPC, // System V IPC / message queues
		// Make mounts in the child private so e.g. mounting /proc
		// doesn't leak back out onto the host.
		Unshareflags: syscall.CLONE_NEWNS,
	}

	// Start (non-blocking) so we can record the host PID before waiting.
	must(cmd.Start())
	writeMeta(id, cmd.Process.Pid, os.Args[2:])
	must(cmd.Wait())

	// Container exited — clean up metadata and cgroup.
	os.RemoveAll(filepath.Join(RUN_DIR, id))
	os.RemoveAll(filepath.Join("/sys/fs/cgroup/gocker", id))
}

// child runs INSIDE the new namespaces. Its job is to finish building the
// sandbox (limits + filesystem + hostname) and then exec the user's command.
func child() {
	// Args layout: child <id> <command> [args...]
	id := os.Args[2]
	fmt.Printf("[child]  container %s  pid %d (PID 1 inside container)\n", id, os.Getpid())

	setupCgroup(id)
	must(syscall.Sethostname([]byte("container")))

	// Pivot into the container's own filesystem.
	must(syscall.Chroot(ROOTFS))
	must(syscall.Chdir("/"))

	// A fresh /proc so tools like `ps` reflect the PID namespace, not the host.
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())

	must(syscall.Unmount("/proc", 0))
}
