package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/model"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := "root:jiongs@tcp(127.0.0.1:3306)/ebidsystem_test?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	prepareSchema(t, db)

	return db
}

func prepareSchema(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS orders (
		id BIGINT PRIMARY KEY,
		quantity BIGINT NOT NULL,
		filled_quantity BIGINT NOT NULL,
		status VARCHAR(20) NOT NULL,
		updated_at DATETIME
	);
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS trades (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		event_id VARCHAR(64) UNIQUE,
		buy_order_id BIGINT,
		sell_order_id BIGINT,
		price DECIMAL(18,2),
		quantity BIGINT
	);
	`)

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS match_events (
		event_id VARCHAR(64) PRIMARY KEY,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)

	require.NoError(t, err)
}

func TestHandleMatchEvent_TransactionalRollback(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepo(db)
	ctx := context.Background()

	// 清理
	db.Exec("DELETE FROM trades")
	db.Exec("DELETE FROM orders")

	// 插入买单
	res, err := db.Exec(`
		INSERT INTO orders (id, quantity, filled_quantity, status)
		VALUES (1, 100, 0, 'pending')
	`)
	require.NoError(t, err)
	_, _ = res.RowsAffected()

	err = repo.WithTx(ctx, func(tx *sql.Tx) error {
		// 买单成功
		err := repo.FillOrderTx(ctx, tx, 1, 10)
		require.NoError(t, err)

		// 卖单不存在 → 失败
		return repo.FillOrderTx(ctx, tx, 999, 10)
	})

	require.Error(t, err)

	// 验证买单未被修改
	var filled int
	err = db.QueryRow(
		"SELECT filled_quantity FROM orders WHERE id = 1",
	).Scan(&filled)
	require.NoError(t, err)
	require.Equal(t, 0, filled)

	// 验证无成交
	var cnt int
	db.QueryRow("SELECT COUNT(*) FROM trades").Scan(&cnt)
	require.Equal(t, 0, cnt)
}

func TestHandleMatchEvent_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepo(db)
	ctx := context.Background()

	db.Exec("DELETE FROM trades")
	db.Exec("DELETE FROM orders")

	// 买卖单
	db.Exec(`
		INSERT INTO orders (id, quantity, filled_quantity, status)
		VALUES
		(1, 100, 0, 'pending'),
		(2, 100, 0, 'pending')
	`)

	eventID := "evt-123"

	apply := func() error {
		return repo.WithTx(ctx, func(tx *sql.Tx) error {

			ok, err := repo.TryInsertMatchEventTx(ctx, tx, eventID)
			if err != nil {
				return err
			}
			if !ok {
				// 已处理过，直接返回
				return nil
			}

			if err := repo.FillOrderTx(ctx, tx, 1, 10); err != nil {
				return err
			}
			if err := repo.FillOrderTx(ctx, tx, 2, 10); err != nil {
				return err
			}

			trade := &model.Trade{
				EventID:     eventID,
				BuyOrderID:  1,
				SellOrderID: 2,
				Price:       100,
				Quantity:    10,
			}
			return repo.CreateTradeTx(ctx, tx, trade)
		})
	}

	require.NoError(t, apply())
	require.NoError(t, apply()) // 重复执行

	// 验证只成交一次
	var cnt int
	db.QueryRow("SELECT COUNT(*) FROM trades").Scan(&cnt)
	require.Equal(t, 1, cnt)

	// 验证 filled_quantity 只增加一次
	var filled1, filled2 int
	db.QueryRow("SELECT filled_quantity FROM orders WHERE id=1").Scan(&filled1)
	db.QueryRow("SELECT filled_quantity FROM orders WHERE id=2").Scan(&filled2)

	require.Equal(t, 10, filled1)
	require.Equal(t, 10, filled2)
}
