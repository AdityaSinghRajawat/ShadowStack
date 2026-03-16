# ShadowStack 🕵️‍♂️

ShadowStack is a high-performance, language-agnostic SQL profiler powered by eBPF and Go. It provides deep, real-time visibility into database traffic by intercepting queries at the Linux kernel level.

Unlike traditional profilers, ShadowStack requires zero code changes, no library imports, and works across any tech stack (Python, Node.js, Go, Java, etc.) with near-zero performance overhead.

## ✨ Key Features

- **Zero-Code Instrumentation**: Monitor any application without touching a single line of source code.
- **eBPF Kernel-Level Tap**: Intercepts SQL packets (Postgres/MySQL) directly in the kernel for maximum efficiency.
- **Language Agnostic**: Works universally across all programming languages and frameworks.
- **Interactive TUI Dashboard**: A reactive terminal interface built with Bubble Tea for real-time query visualization.
- **Low Overhead**: Uses eBPF Ring Buffers to ensure your production application's performance remains unaffected.

## 🏗 Architecture

ShadowStack splits logic between the high-performance Linux Kernel and a reactive Go userspace.

┌─────────────────────────────────────────────────────────────────────────────┐
│ Linux Kernel │
│ │
│ ┌────────────────────────┐ ┌──────────────────────────────────┐ │
│ │ eBPF Socket Filter │ ──────▶ │ eBPF Ring Buffer │ │
│ │ (shadowStack.c) │ │ (High-speed event transport) │ │
│ └────────────────────────┘ └──────────────────────────────────┘ │
└───────────────────────────────────────────────│─────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Userspace (Go) │
│ │
│ ┌────────────────────────┐ ┌──────────────────────────────────┐ │
│ │ eBPF Loader │ ──────▶ │ Wire-Protocol Parsers │ │
│ │ (internal/ebpf) │ │ (internal/protocol) │ │
│ └────────────────────────┘ └──────────────────────────────────┘ │
│ │ │
│ ▼ │
│ ┌──────────────────────────────────┐ │
│ │ Interactive TUI (UI) │ │
│ │ (internal/ui) │ │
│ └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘

## 📂 Project Structure

```text
shadowStack/
├── bpf/                   # Kernel-space C code (eBPF logic)
├── cmd/                   # Application entry point
├── internal/              # Private modules
│   ├── ebpf/              # Loader & Ring Buffer handlers
│   ├── ui/                # Terminal UI (Bubble Tea)
│   └── protocol/          # SQL Wire-protocol parsers
├── Makefile               # Build & bpf2go automation
├── go.mod                 # Dependency management
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- **OS**: Linux (Kernel 5.4+) or WSL2 / Docker Desktop
- **Tooling**: Go 1.21+, Clang, LLVM, and libbpf-dev
- **Privileges**: Root/Sudo access required to load eBPF probes

### Installation & Build

1. **Clone the repository**

   ```bash
   git clone https://github.com/yourusername/shadowStack.git
   cd shadowStack
   ```

2. **Generate eBPF Bindings & Build**
   ```bash
   # Generates Go code from C and compiles the binary
   make build
   ```

### Running

Monitor PostgreSQL traffic (Port 5432) instantly:

```bash
sudo ./shadowStack --port 5432
```

## 🛠 Usage

- **Navigation**: Use arrow keys to scroll through the live query list.
- **Filtering**: Press `/` to search for specific table names or SQL keywords.
- **Detail View**: Press `Enter` on a query to see the full raw payload and latency breakdown.
- **Exit**: Press `Ctrl+C` or `q` to quit.

## 🤝 Contributing

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.
