# AtlasVM — A Distributed Virtual Machine

AtlasVM is an educational distributed virtual machine built from scratch in Go. It abstracts the full stack of computation: compiling a custom C-like language (AtlasPL) into bytecode, executing it via a custom VM architecture, and distributing the resulting execution state across a peer-to-peer network to reach decentralized consensus using PBFT (Practical Byzantine Fault Tolerance).

> **Read the full step-by-step deep dive and documentation here:**  
[AtlasVM: Building a Distributed Virtual Machine - Lessons in Compilers and Distributed Systems](https://hmzelidrissi.ma/blog/atlasvm-building-a-distributed-virtual-machine/)

---

## Features

- **AtlasPL Compiler:** A built-from-scratch Lexer, Pratt Parser, and Bytecode Compiler for a custom C-like language.
- **Custom VM Architecture:** 
  - 1024-byte segmented memory (Data / Code)
  - Program Counter (PC) and Accumulator (ACC) registers
  - A 15-instruction opcode set (built around a classic Fetch-Decode-Execute cycle)
- **Distributed PBFT Consensus:** A full-mesh network of nodes using Protocol Buffers and gRPC that securely vote on the final execution memory footprint to guarantee fault-tolerant agreement.

## How to Run It

**Prerequisites:** Go 1.22+

```bash
# Clone the repository
git clone https://github.com/HMZElidrissi/atlas-virtual-machine
cd atlas-virtual-machine

# Build the project
make build

# Run an example program (this will spin up 3 nodes to reach consensus)
./atlasvm examples/even_odd.atlas

# Run a program locally (skip the distributed network consensus)
./atlasvm --local examples/sum.atlas
```

### Included Examples

The `examples/` directory contains ready-to-run `.atlas` programs:

| Command | What it does | Expected Output |
|---|---|---|
| `./atlasvm examples/even_odd.atlas` | Is 10 even or odd? | `0` (even) |
| `./atlasvm examples/sum.atlas` | Calculate 3 + 4 | `7` |
| `./atlasvm examples/absolute.atlas` | Absolute value of 5 | `5` |
| `./atlasvm examples/max.atlas` | Calculates 4 + 4 | `8` |

## Project Structure

```text
atlas-virtual-machine/
├── cmd/atlasvm/             ← CLI Entry point
├── examples/                ← AtlasPL example programs
├── internal/
│   ├── atlaspl/             ← Source code tokenization, AST parsing, and Bytecode generation
│   ├── network/             ← gRPC Node Handlers and PBFT Consensus State Machine 
│   └── vm/                  ← Memory limits, Registers, Stack, execution engine
├── proto/                   ← Protobuf definitions (gRPC structures)
└── Makefile                 ← Tooling to build, step, protocol generate, and clean
```
