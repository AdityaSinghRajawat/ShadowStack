CLANG ?= clang
CFLAGS := -O2 -g -Wall -Werror $(CFLAGS)

all: generate build

generate:
	go generate ./...

build:
	go build -o shadowStack cmd/shadowStack/main.go

vmlinux:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/vmlinux.h

clean:
	rm -f shadowStack
	rm -f internal/ebpf/shadowstack_bpf*
