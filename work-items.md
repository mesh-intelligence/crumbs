# Work items

## Organize files for maximum separation of concerns

To maximize **Separation of Concerns (SoC)** for an AI-native architecture, we need to decouple the **Intent** (Interfaces), the **Domain Logic** (State Transitions), and the **Infrastructure** (Persistence).

In this model, the AI can work on the SQLite implementation without ever touching the Crumb logic, and vice versa.

### The AI-Optimized Directory Structure

```text
crumbs/
├── cmd/
│   └── cupboard/          # CLI Entrypoint (The "User" of the library)
├── pkg/
│   ├── api/               # Intent: Pure Interfaces (Cupboard, Table)
│   ├── schema/            # Data: Entity Structs (Crumb, Trail, Stash)
│   └── constants/         # Rules: States, LinkTypes, PropertyDefaults
├── internal/
│   ├── persistence/       # Infrastructure: The JSONL/SQLite bridge
│   │   ├── mapping/       # Translation: Entity <-> SQL Row logic
│   │   └── engine/        # Driver: SQLite lifecycle & JSONL flushing
│   └── telemetry/         # Observability: OTel implementation
├── scripts/               # AI Tools: Code gen for boilerplates
└── go.mod

```

---

### 1. Decoupling the "What" from the "How"

By splitting `pkg/api` from `pkg/schema`, you prevent the AI from accidentally introducing side effects.

* **`pkg/api`**: Defines the contract. If the AI is asked to add a `Search` method, it only changes this package.
* **`pkg/schema`**: Contains the "In-Memory" logic. Methods like `Crumb.Pebble()` only change fields in the struct. They **never** know the database exists.

---

### 2. The Persistence Layer (The "Mapping" Pattern)

To keep the SQLite backend from becoming a "God Object," we use a **Mapper** pattern inside `internal/persistence`. This is crucial for AI coding because it limits the context the AI needs to read to fix a bug.

| Component | Responsibility | AI Prompt Scope |
| --- | --- | --- |
| **Engine** | Managing `sql.DB` and file locks for JSONL. | "Optimize the JSONL flush interval." |
| **Mapping** | Converting a `types.Crumb` into a `INSERT INTO crumbs...` statement. | "Add a new column for the `DeletedAt` field." |
| **Table** | Satisfying the `Table` interface by calling the Engine and Mapper. | "Implement a retry logic for the `Get` operation." |

---

### 3. Handling the "JSONL as Source of Truth"

Since Decision 6 mandates JSONL as the canonical store, we isolate the synchronization logic into a **Watcher/Writer** pattern.

* **The Write-Ahead Pattern**: When `Table.Set()` is called, the `internal/persistence/engine` writes to the JSONL file first, then updates the SQLite cache.
* **The Recovery Pattern**: On `Attach()`, a specialized `internal/persistence/loader` reads the JSONL and rebuilds the SQLite memory/temp file.

---

### 4. Constants as a Single Source of Truth

To avoid "Magic Strings" across the codebase (which often confuses LLMs), we centralize all domain rules:

```go
// pkg/constants/states.go
package constants

const (
    StateDraft   = "draft"
    StatePebble  = "pebble"
    LinkChildOf  = "child_of"
)

```

---

### 5. AI-Ready Testing Strategy

To maintain this separation, we organize tests by **Layer**:

1. **Unit Tests (`pkg/schema`)**: Tests state transitions in memory (e.g., `Crumb.Pebble()` sets state correctly).
2. **Integration Tests (`internal/persistence`)**: Tests that `Table.Set()` actually results in a JSONL line and a SQLite row.
3. **End-to-End Tests (`cmd/cupboard`)**: Tests the CLI output.

**Would you like me to generate the `internal/persistence/mapping` logic for the Crumb entity to show how we keep the SQL code separate from the domain struct?**