//go:build ignore

typedef unsigned char __u8;
typedef unsigned short __u16;
typedef unsigned int __u32;
typedef unsigned long long __u64;

#define SEC(name) __attribute__((section(name), used))
#define __uint(name, val) int (*name)[val]

#define BPF_MAP_TYPE_RINGBUF 27

static void *(*bpf_ringbuf_reserve)(void *ringbuf, __u64 size, __u64 flags) = (void *) 131;
static void (*bpf_ringbuf_submit)(void *data, __u64 flags) = (void *) 132;
static __u64 (*bpf_get_current_pid_tgid)(void) = (void *) 14;
static long (*bpf_skb_load_bytes)(const void *skb, __u32 offset, void *to, __u32 len) = (void *) 26;

struct __sk_buff {
    __u32 len;
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24); 
} events SEC(".maps");

struct sql_event {
    __u32 pid;
    __u32 payload_len;
    __u8 payload[512]; 
};

SEC("socket")
int socket_handler(struct __sk_buff *skb) {
    __u32 copy_len = skb->len;

    if (copy_len == 0) return 0;
    
    // Check bounds for the new 512-byte buffer
    if (copy_len > 511) {
        copy_len = 511;
    }

    struct sql_event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
    if (!e) return 0; 

    e->pid = bpf_get_current_pid_tgid() >> 32;
    e->payload_len = copy_len;

    // Verifier hint for the 512-byte limit (0x1FF is 511)
    copy_len &= 0x1FF; 
    if (copy_len > 0) {
        bpf_skb_load_bytes(skb, 0, e->payload, copy_len);
    }

    bpf_ringbuf_submit(e, 0);
    return 0;
}

char _license[] SEC("license") = "GPL";