package persist

import (
	"context"
	"fmt"
)

const defaultWALGoldItemID int32 = 40308

const minGeneratedWALItemObjectID int32 = 500_000_000

const recoverWALItemTransferSQL = `UPDATE character_items SET char_id = $1 WHERE obj_id = $2 AND char_id = $3`

// WALEntry represents one economic write-ahead log entry.
type WALEntry struct {
	TxType     string // "trade", "shop", "auction"
	FromChar   int32
	ToChar     int32
	ItemID     int32
	Count      int32
	EnchantLvl int16
	GoldAmount int64
}

func isWALItemTransfer(wal WALEntry) bool {
	return wal.GoldAmount == 0 && wal.ItemID >= minGeneratedWALItemObjectID && wal.FromChar > 0 && wal.ToChar > 0
}

func isWALStackItemTransfer(wal WALEntry) bool {
	return wal.GoldAmount == 0 && wal.ItemID > 0 && wal.ItemID < minGeneratedWALItemObjectID && wal.Count > 0 && wal.FromChar > 0 && wal.ToChar > 0
}

func walGoldItemID(wal WALEntry) int32 {
	if wal.ItemID > 0 {
		return wal.ItemID
	}
	return defaultWALGoldItemID
}

type WALRepo struct {
	db *DB
}

func NewWALRepo(db *DB) *WALRepo {
	return &WALRepo{db: db}
}

// WriteWAL atomically writes a batch of WAL entries in a single transaction.
// Returns nil on success. If it fails, the caller should cancel the operation.
func (r *WALRepo) WriteWAL(ctx context.Context, entries []WALEntry) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("wal begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, e := range entries {
		if _, err := tx.Exec(ctx,
			`INSERT INTO economic_wal (tx_type, from_char, to_char, item_id, count, enchant_lvl, gold_amount)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			e.TxType, e.FromChar, e.ToChar, e.ItemID, e.Count, e.EnchantLvl, e.GoldAmount,
		); err != nil {
			return fmt.Errorf("wal insert: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// MarkProcessed marks all WAL entries as processed (called during batch flush).
func (r *WALRepo) MarkProcessed(ctx context.Context) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE economic_wal SET processed = TRUE WHERE processed = FALSE`,
	)
	return err
}

// RecoverWAL reads all unprocessed WAL entries and replays them.
// Called once at server startup before the game loop begins.
// Each entry is replayed idempotently:
//   - gold_amount > 0: deduct from from_char, add to to_char
//   - item_id > 0: transfer item ownership from from_char to to_char
//
// After replay, entries are marked processed.
func (r *WALRepo) RecoverWAL(ctx context.Context) (int, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, tx_type, from_char, to_char, item_id, count, enchant_lvl, gold_amount
		 FROM economic_wal WHERE processed = FALSE ORDER BY id`)
	if err != nil {
		return 0, fmt.Errorf("wal recover query: %w", err)
	}
	defer rows.Close()

	var entries []struct {
		id    int64
		entry WALEntry
	}
	for rows.Next() {
		var e struct {
			id    int64
			entry WALEntry
		}
		if err := rows.Scan(&e.id, &e.entry.TxType, &e.entry.FromChar, &e.entry.ToChar,
			&e.entry.ItemID, &e.entry.Count, &e.entry.EnchantLvl, &e.entry.GoldAmount); err != nil {
			return 0, fmt.Errorf("wal recover scan: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("wal recover rows: %w", err)
	}

	if len(entries) == 0 {
		return 0, nil
	}

	// Replay each entry in a transaction
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("wal recover begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, e := range entries {
		wal := e.entry

		// Replay gold transfer
		if wal.GoldAmount > 0 && wal.FromChar > 0 && wal.ToChar > 0 {
			goldItemID := walGoldItemID(wal)
			if _, err := tx.Exec(ctx,
				`UPDATE character_items SET count = count - $1::INT WHERE char_id = $2 AND item_id = $3`,
				wal.GoldAmount, wal.FromChar, goldItemID); err != nil {
				return 0, fmt.Errorf("wal recover gold deduct (id=%d): %w", e.id, err)
			}
			if _, err := tx.Exec(ctx,
				`DELETE FROM character_items WHERE char_id = $1 AND item_id = $2 AND count <= 0`,
				wal.FromChar, goldItemID); err != nil {
				return 0, fmt.Errorf("wal recover gold cleanup (id=%d): %w", e.id, err)
			}
			tag, err := tx.Exec(ctx,
				`UPDATE character_items SET count = count + $1::INT WHERE char_id = $2 AND item_id = $3`,
				wal.GoldAmount, wal.ToChar, goldItemID)
			if err != nil {
				return 0, fmt.Errorf("wal recover gold add (id=%d): %w", e.id, err)
			}
			if tag.RowsAffected() == 0 {
				if _, err := tx.Exec(ctx,
					`INSERT INTO character_items (char_id, item_id, count, obj_id) VALUES ($1, $2, $3::INT, 0)`,
					wal.ToChar, goldItemID, wal.GoldAmount); err != nil {
					return 0, fmt.Errorf("wal recover gold insert (id=%d): %w", e.id, err)
				}
			}
		}

		if isWALStackItemTransfer(wal) {
			if _, err := tx.Exec(ctx,
				`UPDATE character_items SET count = count - $1::INT WHERE char_id = $2 AND item_id = $3`,
				wal.Count, wal.FromChar, wal.ItemID); err != nil {
				return 0, fmt.Errorf("wal recover stack item deduct (id=%d): %w", e.id, err)
			}
			if _, err := tx.Exec(ctx,
				`DELETE FROM character_items WHERE char_id = $1 AND item_id = $2 AND count <= 0`,
				wal.FromChar, wal.ItemID); err != nil {
				return 0, fmt.Errorf("wal recover stack item cleanup (id=%d): %w", e.id, err)
			}
			tag, err := tx.Exec(ctx,
				`UPDATE character_items SET count = count + $1::INT WHERE char_id = $2 AND item_id = $3`,
				wal.Count, wal.ToChar, wal.ItemID)
			if err != nil {
				return 0, fmt.Errorf("wal recover stack item add (id=%d): %w", e.id, err)
			}
			if tag.RowsAffected() == 0 {
				if _, err := tx.Exec(ctx,
					`INSERT INTO character_items (char_id, item_id, count, enchant_lvl, obj_id) VALUES ($1, $2, $3::INT, $4, 0)`,
					wal.ToChar, wal.ItemID, wal.Count, wal.EnchantLvl); err != nil {
					return 0, fmt.Errorf("wal recover stack item insert (id=%d): %w", e.id, err)
				}
			}
		}

		// Replay item transfer
		if isWALItemTransfer(wal) {
			if _, err := tx.Exec(ctx, recoverWALItemTransferSQL, wal.ToChar, wal.ItemID, wal.FromChar); err != nil {
				return 0, fmt.Errorf("wal recover item transfer (id=%d): %w", e.id, err)
			}
		}

		// Mark this entry as processed
		if _, err := tx.Exec(ctx,
			`UPDATE economic_wal SET processed = TRUE WHERE id = $1`, e.id); err != nil {
			return 0, fmt.Errorf("wal recover mark (id=%d): %w", e.id, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("wal recover commit: %w", err)
	}

	return len(entries), nil
}
