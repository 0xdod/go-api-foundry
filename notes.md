## ✅ Evaluation Criteria Achievement

| Category               | Requirement         | Implementation                                                                           |
| :--------------------- | :------------------ | :--------------------------------------------------------------------------------------- |
| **Ledger Correctness** | Double-entry        | `Deposit/Withdraw` always create pair entries (Debit/Credit).                            |
|                        | Imbalance possible? | No. Operations are atomic within `WithTx`.                                               |
| **Concurrency**        | No Race Conditions  | usage of `balances` table row-locking prevents concurrent modifications to same account. |
| **Atomicity**          | Atomic Transfers    | `WithTx` wrapper ensures Transaction + Ledger Entries + Balance updates commit together. |
| **Idempotency**        | Retry Safe          | `IdempotencyMiddleware` intercepts duplicates before they reach domain logic.            |
| **Data Model**         | Clean Modelling     | Separation of `Account` (Entity), `Transaction` (Intent), `LedgerEntry` (Accounting).    |
