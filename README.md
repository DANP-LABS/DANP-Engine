# DANP-Engine

![build](https://github.com/IceFireLabs/DANP-Engine/actions/workflows/build.yml/badge.svg)
![test](https://github.com/IceFireLabs/DANP-Engine/actions/workflows/test.yml/badge.svg)

**DANP-Engine** is a trusted AI MCP runtime built on **IPFS**, **WASM**, **AI MCP Server** and **AI MCP Client**. The AI MCP tools modules are stored on decentralized IPFS immutable trusted storage. This innovative framework seamlessly integrates WebAssembly computing, IPFS storage, and AI Agent capabilities to deliver a robust, serverless solution for decentralized applications (dApps). Designed to empower developers with cutting-edge technologies, this project revolutionizes decentralized computing by combining the efficiency of WASM, the reliability of IPFS, and the intelligence of AI Agents.

![image](https://github.com/user-attachments/assets/45bf263a-1624-49a5-b113-8d7e430a4d18)

---

## Key Features and Components

- **WebAssembly Computing**: Harness the power of WebAssembly for efficient, scalable, and portable computation within decentralized applications.
- **IPFS Integration**: Utilize IPFS for secure, decentralized storage and retrieval of data, ensuring data integrity and availability.
- **AI Agent Capabilities**: Empower your applications with intelligent decision-making, automation, and multi-agent control powered by AI.
- **Serverless Architecture**: Embrace a serverless paradigm for flexible, cost-effective deployment of dApps, eliminating the complexities of traditional server management.
- **Technical Support**: Benefit from seamless integration with **Filecoin-Lassie** for IPFS file retrieval, **Filecoin-IPLD-Go-Car** for IPFS Car file extraction, **Extism** for WASM plugin management, **wazero** for WASM virtual machine, and **Fiber** for high-performance HTTP server capabilities.

---
## System Architecture

```mermaid
graph LR
    %% DANP-Engine Architecture Diagram
    description[IPFS Trusted Storage + WASM Edge Computing + AI Agents]

    %% IPFS Storage Layer
    subgraph ipfs[IPFS Storage Layer]
        direction TB
        ipfs_node1[IPFS Node]
        ipfs_node2[IPFS Node]
        ipfs_node3[IPFS Node]
        ipfs_node1 <--> ipfs_node2
        ipfs_node2 <--> ipfs_node3
    end

    %% WASM Runtime Layer
    subgraph runtime[WASM Runtime Layer]
        direction LR
        
        subgraph node1[Runtime Node 1]
            wasm1[WASM Engine]
            agent1[AI Agent]
            wasm1 --> agent1
        end
        
        subgraph node2[Runtime Node 2]
            wasm2[WASM Engine]
            agent2[AI Agent]
            wasm2 --> agent2
        end
        
        subgraph node3[Runtime Node 3]
            wasm3[WASM Engine]
            agent3[AI Agent]
            wasm3 --> agent3
        end
    end

    %% Application Layer
    subgraph apps[Application Layer]
        direction TB
        edge_ai[Edge AI]
        trusted_comp[Trusted Computing]
        edge_func[Edge Functions]
    end

    %% Services Layer
    subgraph services[Services]
        direction LR
        compute[Compute Service]
        storage[Storage Service]
        network[Network Service]
    end

    %% Connections
    ipfs --> runtime
    node1 --> apps
    node2 --> apps
    node3 --> apps
    apps --> services
    
    %% Styling
    style ipfs fill:#e3f9ff,stroke:#333
    style runtime fill:#fff2e6,stroke:#333
    style apps fill:#e6ffe6,stroke:#333
    style services fill:#f9e6ff,stroke:#333
    
    classDef tech fill:#f8f9fa,stroke:#333,stroke-width:1px;
    class description,ipfs_node1,ipfs_node2,ipfs_node3,wasm1,wasm2,wasm3,agent1,agent2,agent3,edge_ai,trusted_comp,edge_func,compute,storage,network tech
```
The architecture consists of four main layers:

* IPFS Storage Layer: Decentralized storage network that provides content-addressable storage for WASM modules and application data
* WASM Runtime Layer: Distributed execution environment where WASM modules run with AI Agent capabilities
* Application Layer: Core functionalities exposed to developers including Edge AI, Trusted Computing and Edge Functions
* Services Layer: Final output services that applications can consume (Compute, Storage, Network)

## How It's Made

Here's a breakdown of how **DANP-Engine** was built, including the technologies used, their integration, and notable aspects:

### WebAssembly (WASM) Computing
- **Technology**: Leveraged WebAssembly for its efficient and portable bytecode format, enabling high-performance computing within the serverless environment.
- **Integration**: Integrated WebAssembly runtime libraries and tools to compile and execute WASM modules seamlessly within the serverless framework.

### IPFS Integration
- **Technology**: Utilized IPFS (InterPlanetary File System) for decentralized storage and retrieval of data.
- **Integration**: Integrated IPFS libraries and APIs to interact with the IPFS network, ensuring secure and decentralized data storage and retrieval.

### AI Agent Capabilities
- **Technology**: Integrated advanced AI frameworks to enable intelligent decision-making, automation, and multi-agent control.
- **Integration**: Combined AI Agent capabilities with WASM and IPFS to create a powerful, intelligent, and decentralized computing ecosystem.

### Partner Technologies and Benefits
- **Filecoin-Lassie**: Leveraged for IPFS file retrieval, enhancing data access capabilities within the serverless environment.
- **Filecoin-IPLD-Go-Car**: Used for IPFS Car file extraction, enabling efficient handling of IPFS Car files within the project.
- **Extism**: Integrated for WASM plugin management, facilitating extensibility and customization of the serverless environment through WASM plugins.
- **wazero**: Utilized for WASM virtual machine capabilities, ensuring efficient execution of WebAssembly code within the serverless framework.
- **Fiber**: Integrated for high-performance HTTP server functionalities, enhancing network communication and HTTP request handling within the serverless environment.

### Notable Aspects
- **Dynamic IPFS Integration**: Dynamically integrated with IPFS using Filecoin-Lassie and IPFS CID references in the configuration, allowing for seamless interaction with IPFS resources.
- **AI-Driven Automation**: Enabled intelligent automation and decision-making through AI Agent integration, making decentralized computing smarter and more efficient.

---

## Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/IceFireLabs/DANP-Engine.git
```

### 2. Build the Project
```bash
cd DANP-Engine
make
```

### 3. Adjust Configuration File
```yaml
# MCP Server Manifest
server_config:
  host: "0.0.0.0"
  port: 18080
  max_connections: 100
  timeout: 30s

ipfs:
  enable: true  # Set to true to enable IPFS support
  lassie_net:
    scheme: "http"  # http or https
    host: "127.0.0.1"
    port: 31999
  cids: []  # Optional list of pre-loaded CIDs

llm_config:
  base_url: ""  # Optional base URL for API endpoints
  provider: "openai"  # Default provider
  openai:  # OpenAI-specific config
    api_key: ""
    model: "gpt-4"
    temperature: 0.7
    max_tokens: 2048
  # Add other provider configs here as needed

# Defines WASM modules and their exposed MCP tools
modules:
  - name: "hello"
    #wasm_path: "file:///home/corerman/ICODE/IceFireLabs/dANP-Engine/config/hello.wasm"  # Supports file:// or IPFS:// schemes
    wasm_path: "IPFS://QmeDsaLTc8dAfPrQ5duC4j5KqPdGbcinEo5htDqSgU8u8Z"  # Supports file:// or IPFS:// schemes
    tools:
      - name: "say_hello"
        description: "Greet someone by name"
        inputs:
          - name: "name"
            type: "string"
            required: true
            description: "Name to greet"
        outputs:
          type: "string"
          description: "Greeting message"

```

### 4. Load Configuration and Run MCP Server
```bash
go run cmd/DANP-MCP-SERVER/main.go
```

### 5. Interact with MCP Server using Client
```bash
go run cmd/DANP-MCP-CLIENT/main.go -http http://localhost:18080/
```

### 6. Example AI Interaction
```bash
# Server startup log showing WASM module loading from IPFS
2025/06/29 14:06:10 Loading WASM module from IPFS CID: QmeDsaLTc8dAfPrQ5duC4j5KqPdGbcinEo5htDqSgU8u8Z
2025/06/29 14:06:10 Successfully loaded WASM module: IPFS://QmeDsaLTc8dAfPrQ5duC4j5KqPdGbcinEo5htDqSgU8u8Z
2025/06/29 14:06:10 Registering tool: say_hello
2025/06/29 14:06:10 MCP server listening on 0.0.0.0:18080

# Client interaction example
Enter your request (empty line to submit, 'exit' to quit):
> Could you please greet my friend John for me?
> 

AI Response:
I've greeted your friend John for you! Here's the message: 

👋 Hello John
```

---

## Contributing

We welcome contributions from the community! To contribute to **DANP-Engine**:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Make your changes and commit them (`git commit -am 'Add new feature'`).
4. Push your changes to the branch (`git push origin feature/your-feature`).
5. Create a new Pull Request.

---

## License

This library is dual-licensed under **Apache 2.0** and **MIT** terms.

---

## ❤️ Thanks for Technical Support ❤️

1. [**Filecoin-Lassie**](https://github.com/filecoin-project/lassie/): Support IPFS file retrieval.
2. [**Filecoin-IPLD-Go-Car**](https://github.com/ipld/go-car): Support IPFS Car file extraction.
3. [**Extism**](https://extism.org/): WASM plugin management.
4. [**wazero**](https://wazero.io/): WASM virtual machine.
5. [**Fiber**](https://gofiber.io/): High-performance HTTP server.
