package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/config"
	"ebidsystem_csm/internal/model"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	if dsn == "" {
		cfg, err := config.Load()
		if err == nil {
			dsn = strings.TrimSpace(cfg.MySQLTest.DSN)
		}
	}
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is empty and config MySQLTest.DSN is unavailable; skip MySQL integration test")
	}

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	prepareSchema(t, db)

	return db
}

func prepareSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	// 事务测试应使用专用测试库。这里重建表结构，保证测试可重复执行。
	_, err := db.Exec(`DROP TABLE IF EXISTS trades`)
	require.NoError(t, err)
	_, err = db.Exec(`DROP TABLE IF EXISTS match_events`)
	require.NoError(t, err)
	_, err = db.Exec(`DROP TABLE IF EXISTS orders`)
	require.NoError(t, err)

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS orders (
		id BIGINT PRIMARY KEY,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		deleted_at DATETIME NULL,
		symbol VARCHAR(64) NOT NULL DEFAULT 'TEST',
		` + "`type`" + ` VARCHAR(16) NOT NULL DEFAULT 'limit',
		side VARCHAR(16) NOT NULL DEFAULT 'buy',
		price DECIMAL(18,4) NULL,
		quantity BIGINT NOT NULL,
		filled_quantity BIGINT NOT NULL DEFAULT 0,
		user_id BIGINT NOT NULL DEFAULT 0,
		status VARCHAR(20) NOT NULL
	);
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS trades (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		event_id VARCHAR(128) NOT NULL,
		symbol VARCHAR(64) NOT NULL,
		buy_order_id BIGINT UNSIGNED NOT NULL,
		sell_order_id BIGINT UNSIGNED NOT NULL,
		price DECIMAL(18,4) NOT NULL,
		quantity BIGINT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY uk_trades_event_id (event_id)
	);
	`)

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS match_events (
		event_id VARCHAR(128) PRIMARY KEY,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`)

	require.NoError(t, err)
}

func cleanTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("DELETE FROM trades")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM match_events")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM orders")
	require.NoError(t, err)
}

func TestFillOrderTx_TransactionalRollback(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepo(db)
	ctx := context.Background()

	cleanTables(t, db) //清理各表

	// 插入买单：
	_, err := db.Exec(`
		INSERT INTO orders (id, quantity, filled_quantity, status)
		VALUES (1, 100, 0, 'pending')
	`)
	require.NoError(t, err)

	err = repo.WithTx(ctx, func(tx *sql.Tx) error {
		// 买单成功（filled_quantity 增加 10）：
		err := repo.FillOrderTx(ctx, tx, "AAPL", 1, 10)
		require.NoError(t, err)
		// 卖单不存在 → 失败（事务回滚，买单的 filled_quantity 不会增加）：
		return repo.FillOrderTx(ctx, tx, "AAPL", 999, 10)
	})

	require.Error(t, err)

	// 验证买单未被修改：
	var filled int
	err = db.QueryRow(
		"SELECT filled_quantity FROM orders WHERE id = 1",
	).Scan(&filled)
	require.NoError(t, err)
	require.Equal(t, 0, filled)

	// 验证无成交：
	var cnt int
	err = db.QueryRow("SELECT COUNT(*) FROM trades").Scan(&cnt)
	require.NoError(t, err)
	require.Equal(t, 0, cnt)
}

func TestHandleMatchEvent_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepo(db)
	ctx := context.Background()

	cleanTables(t, db) //清理各表

	// 买卖单：
	_, err := db.Exec(`
		INSERT INTO orders (id, symbol, quantity, filled_quantity, status)
		VALUES
		(1, 'AAPL', 100, 0, 'pending'),
		(2, 'AAPL', 100, 0, 'pending')
	`)
	require.NoError(t, err)

	eventID := "evt-123"

	apply := func() error {
		return repo.WithTx(ctx, func(tx *sql.Tx) error {
			ok, err := repo.InsertMatchEventTx(ctx, tx, eventID, "AAPL", 1, 2, 10, 100)
			if err != nil {
				return err
			}
			if !ok {
				return nil //已处理过，直接返回
			}

			if err := repo.FillOrderTx(ctx, tx, "AAPL", 1, 10); err != nil {
				return err
			}
			if err := repo.FillOrderTx(ctx, tx, "AAPL", 2, 10); err != nil {
				return err
			}

			trade := &model.Trade{
				EventID:     eventID,
				Symbol:      "AAPL",
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
	err = db.QueryRow("SELECT filled_quantity FROM orders WHERE id=1").Scan(&filled1)
	require.NoError(t, err)
	err = db.QueryRow("SELECT filled_quantity FROM orders WHERE id=2").Scan(&filled2)
	require.NoError(t, err)

	require.Equal(t, 10, filled1)
	require.Equal(t, 10, filled2)
}
