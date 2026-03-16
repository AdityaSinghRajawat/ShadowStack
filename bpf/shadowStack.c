//go:build ignore

// --- 1. Minimal eBPF Headers (Self-Contained for macOS Cross-Compilation) ---
typedef unsigned char __u8;
typedef unsigned short __u16;
typedef unsigned int __u32;
typedef unsigned long long __u64;

#define SEC(name) __attribute__((section(name), used))
#define __uint(name, val) int (*name)[val]

#define BPF_MAP_TYPE_RINGBUF 27

// Manually mapping BPF helper functions to their Linux Kernel instruction IDs
static void *(*bpf_ringbuf_reserve)(void *ringbuf, __u64 size, __u64 flags) = (void *) 131;
static void (*bpf_ringbuf_submit)(void *data, __u64 flags) = (void *) 132;
static __u64 (*bpf_get_current_pid_tgid)(void) = (void *) 14;
static long (*bpf_skb_load_bytes)(const void *skb, __u32 offset, void *to, __u32 len) = (void *) 26;

// The network packet structure
struct __sk_buff {
    __u32 len;
    __u32 pkt_type;
    __u32 mark;
    __u32 queue_mapping;
    __u32 protocol;
    __u32 vlan_present;
    __u32 vlan_tci;
    __u32 vlan_proto;
    __u32 priority;
    __u32 ingress_ifindex;
    __u32 ifindex;
    __u32 tc_index;
    __u32 cb[5];
    __u32 hash;
    __u32 tc_classid;
    __u32 data;
    __u32 data_end;
    __u32 napi_id;
    __u32 family;
    __u32 remote_ip4;
    __u32 local_ip4;
    __u32 remote_ip6[4];
    __u32 local_ip6[4];
    __u32 remote_port;
    __u32 local_port;
    __u32 data_meta;
    __u32 flow_keys;
    __u64 tstamp;
    __u32 wire_len;
    __u32 gso_segs;
    __u32 sk;
    __u32 gso_size;
};

// --- 2. Application Logic ---

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24); // 16 MB buffer
} events SEC(".maps");

struct sql_event {
    __u32 pid;
    __u32 payload_len;
    __u8 payload[256]; 
};

SEC("socket")
int socket_handler(struct __sk_buff *skb) {
    // Only process packets that have data (payload)
    if (skb->len == 0) {
        return 0;
    }

    // Reserve space in the ring buffer
    struct sql_event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
    if (!e) {
        return 0; // Drop if buffer is full
    }

    e->pid = bpf_get_current_pid_tgid() >> 32;
    
    // Calculate how much data to copy (max 256 bytes)
    __u32 copy_len = skb->len;
    if (copy_len > sizeof(e->payload)) {
        copy_len = sizeof(e->payload);
    }
    e->payload_len = copy_len;

    // Load the packet bytes into our event struct
    bpf_skb_load_bytes(skb, 0, e->payload, copy_len);

    // Submit the event to userspace (Go)
    bpf_ringbuf_submit(e, 0);

    return 0; // Return 0 to let the packet continue normally
}

char _license[] SEC("license") = "GPL";