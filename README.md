# GoXfer 
### Client-Sealed, Stateless, Ephemeral, Zero-Trust, End-to-End Encrypted Data Storage System
---

Security-first file transfer system designed around **zero-trust principles**, **client-side encryption**, and **minimal attack surface**.  
Consists of a Go backend and a terminal-first (TUI) client, intentionally avoiding bloated UIs and implicit trust in servers.

---
> This is **not** a cloud storage clone. It is a systems-oriented project focused on cryptography, secure client-server design, and real-world deployment constraints.

#### TUI Demo Video: <a href="https://github.com/Manaswa-S/GoXfer-TUI/releases/download/demo-v1.2/goxfer-tui-v1-demo.webm">Link</a>
---

### Project Goals & Design Principles

- **Strict Zero-Server-Trust by Design**  
  No plaintext, password-derived secrets, long-term cryptographic material, and assumptions about server honesty.

- **Layered Cryptography**  
  Data protection achieved through multiple cryptographic layers with key wrapping and separation of duties between keys.

- **PAKE-Based Authentication**  
  OPAQUE for authentication, passwords are never revealed to the server and cannot be brute-forced offline even after a server compromise.
  - Learn more about OPAQUE: <a href="https://blog.cloudflare.com/opaque-oblivious-passwords/">Link</a>

- **Client-Dominated Security Model**  
  All security-critical operations (key derivation, encryption, and decryption) occur exclusively on the client.
 
- **Terminal-First Interface**  
  Focused, scriptable, and deterministic client. Avoids browser complexity and hidden behavior.

---

### Security Model

- **Threats Considered**
  - Malicious or compromised server
  - Database leaks
  - Offline brute-force attacks
  - Network interception (MITM)
  - Replay attempts

- **Security Guarantees**
  - Server never sees plaintext files
  - Server never stores password-equivalent secrets
  - Encrypted data is useless if exfiltrated
  - Authentication resists offline attacks

---

### Overview

- **Cryptography Overview**
  - **Client-Side Encryption**
    - Files are encrypted before leaving the client
  - **Key Hierarchy**
    - Content Encryption Keys (CEKs)
    - Key Encryption Keys (KEKs)
    - Multiple wrapping layers
  - **No Plaintext Keys at Rest**
  - **Key Lifecycle Management**
    - Creation -> Use -> Zeroization
  
  The design prioritizes **removal of trust**, not just stronger algorithms.
---

- **Authentication & Authorization**
  - Session-based access with replay protection
  - Clear separation between authentication and encryption keys
---

- **Backend Responsibilities**
  - Authentication verification
  - Secure API endpoints
  - Encrypted file and metadata storage
  - Integrity validation
  - Abuse prevention and rate limiting
  - Logging without leaking sensitive data
  
  The backend treats all stored data as **opaque blobs**.
---

- **TUI Responsibilities**
  - Secure credential input
  - Local key derivation and storage
  - Encryption / decryption
  - Upload and download orchestration
  - Progress tracking and retries
  - Atomic file handling on disk

  The TUI is designed to be fast, minimal, and distraction-free.
---

- **Failure Handling & Reliability**
  - Chunked transfers
  - Idempotent operations where possible
---

- **Performance Considerations**
  - Controlled memory usage
  - Efficient Go concurrency
  - Minimal serialization overhead
  
  Security is enforced without turning the system slow or fragile.
---

### Future Roadmap

- **Per-File OPAQUE Authentication**  
  OPAQUE-based authentication to individual file secrets, ensuring metadata and access are revealed only after complete authentication.

- **Adaptive Rate Limiting**  
  State-aware and behavior-based rate limiters to mitigate abuse without relying on naive request counting.

- **Sharded Storage with Server-Side Wrapping**  
  Storage sharding and an additional server-level key wrapping layer to reduce blast radius.

---

### Disclaimer

This project is experimental and educational.  
It has not been professionally audited.  
Use at your own risk.
