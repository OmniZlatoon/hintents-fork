Erst By Hintents
Erst is a premium developer toolset for the Stellar network, designed to provide high-fidelity "glass-box" debugging and simulation for Soroban smart contracts.

Status: Active Development (Phase 4: Advanced Diagnostics) Documentation: https://dotandev-hintents-75.mintlify.app/ Focus: High-Fidelity Simulation, Auth Tracing, and Security Auditing

Scope & Objective
The primary goal of erst is to eliminate the opaque "black box" experience of failed Stellar smart contract transactions. By providing local-first, high-fidelity replay and tracing, erst maps generic network errors back to human-readable diagnostic events and source code.

Core Features (Planned):

Transaction Replay: Fetch a failed transaction's envelope and ledger state from an RPC provider.
Local Simulation: Re-execute the transaction logically in a local environment.
Trace decoding: Map execution steps and failures back to readable instructions or Rust source lines.
Source Mapping: Map WASM instruction failures to specific Rust source code lines using debug symbols.
GitHub Source Links: Automatically generate clickable GitHub links to source code locations in traces (when in a Git repository).
Error Suggestions: Heuristic-based engine that suggests potential fixes for common Soroban errors.
Usage (MVP)
Debugging a Transaction
Fetches a transaction envelope from the Stellar Public network and prints its XDR size (Simulation pending).

./erst debug <transaction-hash> --network testnet
Debug an offline envelope from stdin (no RPC):

./erst debug < tx.xdr
Interactive Trace Viewer
Launch an interactive terminal UI to explore transaction execution traces with search functionality.

./erst debug <transaction-hash> --interactive
# or
./erst debug <transaction-hash> -i
Features:

Search: Press / to search through traces (contract IDs, functions, errors)
Help overlay: Press ? or h to see all keyboard shortcuts
Tree Navigation: Expand/collapse nodes, navigate with arrow keys
Syntax Highlighting: Color-coded contract IDs, functions, and errors
Fast Navigation: Jump between search matches with n/N
Match Counter: See "Match 2 of 5" status while searching
See internal/trace/README.md for detailed documentation.

Performance Profiling
Generate interactive flamegraphs to visualize CPU and memory consumption during contract execution:

./erst debug --profile <transaction-hash>
This generates an interactive HTML file (<tx-hash>.flamegraph.html) with:

Hover tooltips showing frame details (function name, duration, percentage)
Click-to-zoom to focus on specific call stacks
Search/highlight to find frames by name
Dark mode support that adapts to your system theme
Export Formats:

# Interactive HTML (default)
./erst debug --profile --profile-format html <transaction-hash>

# Raw SVG with dark mode support
./erst debug --profile --profile-format svg <transaction-hash>
See docs/INTERACTIVE_FLAMEGRAPH.md for detailed documentation and docs/examples/sample_flamegraph.html for a live demo.

Audit log signing (software / HSM)
erst includes a small utility command to generate a deterministic, signed audit log from a JSON payload.

Software signing (Ed25519 private key)
Provide a PKCS#8 PEM Ed25519 private key via env or CLI:

Env: ERST_AUDIT_PRIVATE_KEY_PEM
CLI: --software-private-key <pem>
Example:

node dist/index.js audit:sign \
  --payload '{"input":{},"state":{},"events":[],"timestamp":"2026-01-01T00:00:00.000Z"}' \
  --software-private-key "$(cat ./ed25519-private-key.pem)"
PKCS#11 HSM signing
Select the PKCS#11 provider with --hsm-provider pkcs11 and configure the module/token/key via env vars.

Required env vars:

ERST_PKCS11_MODULE (path to the PKCS#11 module .so)
ERST_PKCS11_PIN
ERST_PKCS11_KEY_LABEL or ERST_PKCS11_KEY_ID (hex)
ERST_PKCS11_PUBLIC_KEY_PEM (SPKI PEM public key for verification/audit metadata)
Optional:

ERST_PKCS11_SLOT (numeric index into the slot list)
ERST_PKCS11_TOKEN_LABEL
ERST_PKCS11_PIV_SLOT (YubiKey PIV slot such as 9a, 9c, 9d, 9e, 82-95, f9)
Example:

export ERST_PKCS11_MODULE=/usr/lib/softhsm/libsofthsm2.so
export ERST_PKCS11_PIN=1234
export ERST_PKCS11_KEY_LABEL=erst-audit-ed25519
export ERST_PKCS11_PUBLIC_KEY_PEM="$(cat ./ed25519-public-key-spki.pem)"

node dist/index.js audit:sign \
  --hsm-provider pkcs11 \
  --payload '{"input":{},"state":{},"events":[],"timestamp":"2026-01-01T00:00:00.000Z"}'
The command prints the signed audit log JSON to stdout so it can be redirected to a file.

For YubiKey PIV (YKCS11) usage details, see docs/pkcs11-yubikey.md.

Protocol Handler
Erst registers a custom erst:// URI scheme, allowing external tools (browsers, dashboards) to deep-link directly into a debug session.

Register the protocol handler:

./erst protocol:register
Open a debug session via URI:

./erst protocol:handle "erst://debug/<transaction-hash>?network=testnet"
With an optional operation index:

./erst protocol:handle "erst://debug/<transaction-hash>?network=mainnet&op=0"
Unregister the handler when no longer needed:

./erst protocol:unregister
Documentation
Architecture Overview: Deep dive into how the Go CLI communicates with the Rust simulator, including data flow, IPC mechanisms, and design decisions.
Project Proposal: Detailed project proposal and roadmap.
Source Mapping: Implementation details for mapping WASM failures to Rust source code.
Debug Symbols Guide: How to compile Soroban contracts with debug symbols.
Error Suggestions: Heuristic-based error suggestion engine for common Soroban errors.
Canonical JSON Serialization: Deterministic JSON serialization for audit log hashing.
Interactive Trace Showcase: Try out the interactive trace explorer online.
Time Travel Guide: How to use Magic Rewind to replay transactions across time, save sessions to disk, and share reproducible debug files.
Technical Analysis
The Challenge
Stellar's soroban-env-host executes WASM. When it traps (crashes), the specific reason is often sanitized or lost in the XDR result to keep the ledger size small.

The Solution Architecture
erst operates by:

Fetching Data: Using the Stellar RPC to get the TransactionEnvelope and LedgerFootprint (read/write set) for the block where the tx failed.
Simulation Environment: A Rust binary (erst-sim) that integrates with soroban-env-host to replay transactions.
Execution: Feeding the inputs into the VM and capturing diagnostic_events.
For a detailed explanation of the architecture, see docs/architecture.md.

How to Contribute
We are building this open-source to help the entire Stellar community. All contributions, from bug reports to new features, are welcome. Please follow our guidelines to ensure code quality and consistency.

Prerequisites
Go 1.24.0+
Rust 1.70+ (for building the simulator binary)
Stellar CLI (for comparing results)
make (for running standard development tasks)
Getting Started
Clone the repo:

git clone https://github.com/dotandev/hintents.git
cd hintents
Install dependencies:

go mod download
cd simulator && cargo fetch && cd ..
Build the Rust simulator:

cd simulator
cargo build --release
cd ..
Run tests:

go test ./...
cargo test --release -p erst-sim
Development
Code Quality & Linting
This project enforces strict linting rules to maintain code quality. See docs/STRICT_LINTING.md for details.

Quick commands:

# Run all strict linting (Go + Rust)
make lint-all-strict

# Go linting only
make lint-strict

# Rust linting only
make rust-lint-strict

# Install pre-commit hooks
pip install pre-commit && pre-commit install
The CI pipeline fails immediately on:

Unused variables, imports, or functions
Dead code
Any linting warnings
Code Standards
Go Code Style
Formatting: Run go fmt ./... before committing
Linting: Must pass golangci-lint without errors:
golangci-lint run ./...
Naming Conventions:
Use PascalCase for exported identifiers (types, functions, constants)
Use camelCase for unexported identifiers
Use UPPER_SNAKE_CASE for constants
Interface names should end with -er: Reader, Writer, Logger
Error Handling:
Always check and handle errors explicitly
Wrap errors with context using fmt.Errorf: fmt.Errorf("operation failed: %w", err)
Never use bare panic() in production code
Documentation:
All exported functions and types must have documentation comments
Comments should be complete sentences starting with the name
Example: // Logger provides structured logging for diagnostic events.
Rust Code Style
Formatting: Run cargo fmt --all before committing
Linting: Must pass cargo clippy:
cargo clippy --all-targets --release -- -D warnings
Naming Conventions:
Use snake_case for functions and variables
Use PascalCase for types and traits
Use UPPER_SNAKE_CASE for constants
Error Handling:
Prefer Result<T, E> over panics
Use custom error types for domain-specific errors
Avoid unwrapping in production code except for obvious invariants
Documentation:
Document all public functions with doc comments (///)
Include examples for complex functions
Use cargo doc --open to review generated documentation
Commit Message Convention
Follow the Conventional Commits specification:

<type>(<scope>): <subject>

<body>

<footer>
Types:

feat: A new feature
fix: A bug fix
test: Adding or improving tests
docs: Documentation changes
refactor: Code refactoring without feature changes
perf: Performance improvements
chore: Build, CI, or dependency updates
ci: CI/CD configuration changes
Scopes: Use specific areas like sim, cli, updater, trace, analyzer, etc.

Examples:

feat(sim): Add protocol version spoofing for harness
test(sim): Add 1000+ transaction regression suite
fix(updater): Handle network timeouts gracefully
docs: Add comprehensive contribution guidelines
Rules:

Keep subject line under 50 characters
Use imperative mood ("add", not "added" or "adds")
No period at the end of the subject
Provide detailed explanation in the body if the change is non-obvious
Reference related issues: Closes #350, refs #343
Pull Request Structure
Title: Follow commit message convention (this becomes the squashed commit)
Description:
Brief summary of changes
Link to related issues: Closes #XXX
Explain the "why" behind the changes
Highlight any breaking changes
PR Checks:
All CI checks must pass
Code coverage must not decrease
All tests must pass locally before submitting
Format:
## Description
Brief explanation of the changes.

## Related Issues
Closes #350, relates to #343

## Testing
How was this tested? Include specific test cases.

## Checklist
- [ ] Code follows style guidelines
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No new warnings or errors
Testing Requirements
Unit Tests: All new functions must have unit tests
Coverage: Aim for 80%+ coverage. Critical paths should have 90%+ coverage
Integration Tests: Include tests that verify feature interactions
Running Tests:
# Go tests
go test -v -race ./...
go test -v -race -cover ./...

# Rust tests
cargo test --all
cargo test --all --release
Bench Tests: For performance-critical code, include benchmarks
go test -bench=. -benchmem ./...
Development Workflow
Create a branch:

git checkout -b feat/my-feature
# or for bug fixes:
git checkout -b fix/issue-description
Make changes and test locally:

go test ./...
go fmt ./...
golangci-lint run ./...
cargo clippy --all-targets -- -D warnings
cargo fmt --all
Commit with conventional messages:

git add .
git commit -m "feat(scope): description"
Push and create PR:

git push origin feat/my-feature
# Then create PR on GitHub with detailed description
Address feedback:

Make requested changes
Commit with descriptive messages
Force-push if necessary: git push -f origin feat/my-feature
Linting and Formatting
Run the provided scripts before submitting:

# Format Go code
go fmt ./...

# Run linters
golangci-lint run ./...

# Format Rust code
cargo fmt --all

# Check Rust with clippy
cargo clippy --all-targets --release -- -D warnings

# Run all checks
make lint
make format
Development Roadmap
See docs/proposal.md for the detailed proposal.

[x] Phase 1: Research RPC endpoints for fetching historical ledger keys.
[x] Phase 2: Build a basic "Replay Harness" that can execute a loaded WASM file.
[x] Phase 3: Connect the harness to live mainnet data.
[ ] Phase 4: Advanced Diagnostics & Source Mapping (Current Focus).
Common Development Tasks
Running a single test
go test -run TestName ./package
Profiling a test
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
go tool pprof cpu.prof
Building for a specific OS
GOOS=linux GOARCH=amd64 go build -o erst-linux-amd64 ./cmd/erst
Cleaning build artifacts
go clean
cargo clean
make clean
Code Review Checklist
When reviewing PRs, ensure:

[ ] Code follows naming and style conventions
[ ] Error handling is appropriate
[ ] Tests are adequate and pass
[ ] Documentation is clear and complete
[ ] No unnecessary dependencies added
[ ] Performance implications considered
[ ] Security implications reviewed
[ ] Commit messages follow convention
Getting Help
Questions? Open a GitHub Discussion
Found a bug? Create an Issue with reproduction steps
Have an idea? Start a Discussion before implementing
Documentation issue? Create an Issue with details
Important Guidelines
No Emojis: Commit messages and PR titles should not contain emojis
No "Slops": Avoid vague language like "fixes stuff" or "updates things"
Clear Messages: Every commit should have a clear, descriptive message
Lint-Free: Only suppress linting errors if they are objectively false positives. Always explain suppression with // nolint:rule-name comments
Assume Bad Faith in Code: Write code defensively, validate inputs, handle edge cases

[Static Analysis] WASM Semantic Vulnerability Scanner
Repo Avatar
dotandev/hintents
Build a static analysis pass for Soroban WASM binaries that detects high-level semantic vulnerabilities. This includes checking for unchecked host function return values, potential integer overflows in common balance math patterns, and reentrancy-like patterns in cross-contract calls.
Files: internal/analysis/scanner.go [NEW], internal/wat/disassembler.go

I am working on a project with a team of software engineers on an open source project. I want you to take the role of a web developer with more than 15 years of experience to help me on my assingment that is required for the project. You are to execute only what required in this assignment of this project. You must provide a step-by-step process for me to test that i have successfully completed my assignment. Remember, you a web developer with more than 15 years of experience so as to help me on my assingment that is required for the project and you are to execute only what required in this assignment of this project.
