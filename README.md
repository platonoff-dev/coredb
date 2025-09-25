# coredb

coredb is a long-term learning project to build a tiny SQL database from the ground up in Go. The codebase starts with the fundamentals—a durable key-value store—and will eventually layer a relational query engine on top.

## Why two layers?
- **KV storage**: provides an append-friendly, page-based storage engine that any future features can rely on.
- **SQL engine**: will translate SQL statements into KV operations, opening the door to relational features without hard-coding storage details into the query layer.

Keeping these responsibilities separate mirrors the architecture of many production databases and makes it possible to evolve or reuse the KV engine independently.

## Architecture at a glance
### KV storage layer
- `pager.PageManager`: abstracts disk IO into fixed-size pages with `Read`, `Write`, and `Allocate` operations.
- `mem.MemEngine`: an in-memory engine that satisfies the KV interface and supports basic `Get`, `Put`, and `Delete` tests.
- `heap` engine (work-in-progress): sketches out how on-disk heap pages, pointer pages, and iterators will organize records.

### SQL engine layer (planned)
- Will sit on top of the KV primitives and provide table catalogs, query planning, and SQL execution.
- No SQL-facing code lives in the repository yet—the current focus is making the KV layer rock solid.

## Current status
- ✅ Foundational pager abstraction with first-pass tests.
- ✅ In-memory KV engine for rapid prototyping.
- 🚧 Disk-backed heap engine stubs and binary page formats still to implement.
- 🚧 SQL parsing, planning, and execution not started.

## Getting started
1. Install Go 1.24 or newer.
2. Clone the repository and run the tests:
   ```bash
   go test ./...
   ```
3. Explore the `kv/` package to see how storage concerns are being modeled before moving up the stack.

## Roadmap ideas
- Flesh out the heap engine: binary page format, record layout, and free-space management.
- Add a write-ahead log to protect against crashes.
- Design a simple SQL front end (parser, planner, executor) that compiles into KV operations.
- Provide a `coredb` CLI for interacting with the database file.

## Contributing / learning notes
This is primarily a personal sandbox to learn database internals, so the emphasis is on experimenting, reflecting, and iterating slowly. Ideas, suggestions, or references are still very welcome if you are exploring similar territory.
