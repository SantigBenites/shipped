# gocker

A minimal Docker-like container runtime in ~170 lines of Go. Built to show exactly how containers work at the kernel level — no magic, no abstraction layers.

## How containers actually work

Three Linux primitives are all you need:

| Primitive | What it does |
|---|---|
| **Namespaces** | Control what the process can *see* — its own PID tree, hostname, network, mounts |
| **cgroups v2** | Control what the process can *use* — memory, CPU, process count |
| **chroot** | Give the process its own root filesystem |

gocker wires all three together in plain Go.

## Requirements

- Linux (kernel 5.2+ for cgroups v2)
- Root (`sudo`)
- A root filesystem to chroot into (see setup below)

Does **not** work on macOS or Windows — these are real kernel features.

## Setup

**1. Get a root filesystem (Alpine Linux ~3 MB):**

```sh
mkdir -p /tmp/rootfs
cd /tmp/rootfs
curl -fsSL https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.0-x86_64.tar.gz | tar -xz
```

**2. Build gocker:**

```sh
go build -o gocker .
```

## Usage

```sh
# Start an interactive shell inside a container
sudo ./gocker run /bin/sh

# Run a one-shot command
sudo ./gocker run /bin/echo hello from inside

# List running containers (from another terminal)
sudo ./gocker ps
```

### Example session

```
$ sudo ./gocker run /bin/sh
[parent] container a3f19c2e  pid 12345  launching child in new namespaces
[child]  container a3f19c2e  pid 1 (PID 1 inside container)
/ # hostname
container
/ # ps
PID   USER     TIME  COMMAND
    1 root      0:00 /bin/sh
    6 root      0:00 ps
/ # exit
```

```
$ sudo ./gocker ps
CONTAINER ID  PID       CREATED               COMMAND
a3f19c2e      12346     2024-06-07 14:32:01   /bin/sh
```

## Resource limits

Each container is capped at:
- **100 MB** of memory
- **20 processes** (fork-bomb guard)

These are enforced by cgroup v2 and cannot be bypassed from inside the container.

## Code layout

```
main.go    — entry point, command dispatch, must() helper
run.go     — run() and child(): container lifecycle and namespace setup
cgroup.go  — setupCgroup(): per-container resource limits
meta.go    — constants, newID(), writeMeta(): container ID and state
ps.go      — ps(): list running containers
```

### The re-exec trick

Go doesn't have `fork()`. To get a child process into new namespaces, gocker re-executes its own binary (`/proc/self/exe`) with `Cloneflags` set on `SysProcAttr`. The parent calls `gocker child <id> <cmd>` and the kernel creates fresh namespaces for it automatically.

## Limitations

- **No image management** — you point it at a pre-extracted rootfs directory.
- **No networking** — containers start with a loopback-only network namespace.
- **Single rootfs** — all containers share the same `/tmp/rootfs`. For isolation between containers, layer-based copy-on-write (like overlayfs) would be needed.
- **Linux only** — namespaces and cgroups are Linux kernel features.
