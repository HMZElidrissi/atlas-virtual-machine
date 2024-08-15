# AtlasVM

AtlasVM is a distributed virtual machine capable of executing programs written in AtlasPL, a basic programming language. This project explores the challenges and benefits of distributed computing through a simplified virtual machine framework.

## Overview

AtlasVM operates in a distributed system where multiple nodes collaborate to run a program and reach a consensus on the final output. The project consists of several key components:

1. **AtlasPL**: A simple programming language for writing programs to be executed on AtlasVM.
2. **Lexer and Parser**: For analyzing and parsing AtlasPL code.
3. **Interpreter**: For executing AtlasPL programs.
4. **Virtual Machine**: The core component that executes the interpreted code.
5. **Network Layer**: Enables communication between distributed nodes.
6. **Consensus Mechanism**: Ensures agreement on the program's final state across all nodes.

## Getting Started

### Prerequisites

- Go 1.16 or later
- Protocol Buffers compiler (protoc)
- gRPC

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/atlas-virtual-machine.git
   cd atlas-virtual-machine
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Generate Protocol Buffers code:
   ```
   protoc --go_out=. --go-grpc_out=. proto/atlas.proto
   ```

### Running the AtlasVM

1. Start the AtlasVM nodes:
   ```
   go run cmd/atlasvm/main.go
   ```

2. This will start three nodes by default, initialize the network, and execute a sample AtlasPL program.

## Writing AtlasPL Programs

AtlasPL is a simple language supporting basic operations. Here's an example program that checks if a number is even:

```
@ This program checks if a number is even.
var number: int;
number = 10; @ Assign a value to number
if ((number & 1) == 0) { @ Check if last bit is 0 (even)
  return (0);
} else {
  return (1);
}
```

For more details on this project, refer to my blog post on [AtlasVM: Building a Distributed Virtual Machine - Lessons in Compilers and Distributed Systems](https://hmzelidrissi.ma/blog/AtlasVM-Building-a-Distributed-Virtual-Machine).

## Contributing

AtlasVM is for educational purposes and welcomes contributions from the community.
