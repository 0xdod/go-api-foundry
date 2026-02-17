# Go API Foundry - Transaction Ledger System

This project implements a backend API with a **Double-Entry Ledger System** for managing financial transactions (Deposits, Withdrawals, Transfers). It guarantees strict consistency, idempotency, and concurrency safety using PostgreSQL and Go.

## 📚 Ledger Design

The core of the financial system is a **Double-Entry Ledger**. This means that _every_ transaction record is backed by at least two ledger entries (a Debit and a Credit) that sum to zero (implicitly, via the accounting equation).

### Key Features

- **Immutability**: Ledger entries are insert-only. Balances are derived or snapshot-updated, but history is never rewritten.
- **Atomicity**: All operations (Transaction + Ledger Entries + Balance Update) occur within a single ACID database transaction.
- **Correctness**: A transaction cannot be completed without corresponding ledger entries.

### Data Model

The system uses three primary tables:

1.  **`accounts`**: Stores account metadata (ID, Type, Currency).
    - `normal_balance`: Defines if the account is `debit` (Asset/Expense) or `credit` (Liability/Equity) normal.
    - Typical Setup:
      - User Accounts: `credit` normal (Liability from system perspective).
      - Treasury/System Accounts: `debit` normal (Asset).

2.  **`transactions`**: The intent/record of the operation.
    - `reference`: Unique, user-provided or system-generated ID.
    - `type`: deposit, withdraw, transfer.
    - `status`: pending, completed, failed.

3.  **`ledger_entries`**: The immutable history.
    - `transaction_id`: Link to parent transaction.
    - `account_id`: Owner of this entry.
    - `direction`: `debit` or `credit`.
    - `amount`: Always positive integer (sub-units, e.g., cents, kobo).

4.  **`balances`**: A snapshot table for performance and concurrency control.
    - `account_id`
    - `balance`: The current calculated balance.

## 🛡️ Consistency & Concurrency

### Concurrency Approach

Concurrent access to a single account is handled via **Pessimistic Locking** via the `balances` table to ensure safe concurrent access to a single account.

1.  **Lock Acquisition**: When a transaction starts, it calls `UpdateBalance`.
2.  **Row Lock**: The underlying SQL performs an `INSERT ... ON CONFLICT (account_id) DO UPDATE ...` which acquires a **Row-Level Write Lock** on the balance record of the affected account.
3.  **Serialization**: Any other transaction attempting to modify this account waits until the first one commits or rolls back.
4.  **Snapshot Update**: The balance is re-calculated from the sum of historical ledger entries (ensuring correctness) and updated with the new amount.

### Trade-offs

- **Pros**:
  - **Strict Consistency**: Impossible to spend money you don't have (double-spend protection).
  - **Self-Healing**: `ComputeBalance` logic ensures `balances` table matches ledger history.
- **Cons**:
  - **Contention**: High-frequency transactions on a single account (e.g., a central Treasury hot wallet) will be serialized, limiting throughput for that specific account.
  - **Performance**: Summing ledger entries (`ComputeBalance`) is `O(N)` with history size.

### Idempotency Strategy

Idempotency is handled via **Idempotency Keys** to safely handle retries (e.g., network timeouts) and also a unique transaction reference to prevent duplicate transactions.

1.  **Client Responsibility**: Client sends `Idempotency-Key` header.
2.  **Middleware**:
    - If key exists: Return the _previously stored response_ immediately. Logic is skipped.
    - If key is new: Create a "started" record.
    - If processing succeeds: Update record to "completed" with response body.
    - If processing fails: Delete key (allowing retry) or mark enabled depending on error type.
3.  **Storage**: `idempotency_keys` table in Postgres.
4.  **Improvements**: cronjob to expire/delete keys after 24hours.

## 🚀 How to Run

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make
- Swag
- Sqlc

### Running locally

```bash
# Start infrastructure (Postgres)
make dev-migrate

# Run tests (Unit + Integration)
make integration-test
```

### API Documentation

Swagger documentation is available at `/swagger/docs/index.html` when running in dev mode.

## 🧠 Assumptions

- **Currency**: All amounts are integers in lowest denomination (e.g., kobo). Currency conversion is out of scope.
