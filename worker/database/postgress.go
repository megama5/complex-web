package database

import (
	"context"
	"database/sql"
	"errors"
	"local/complex-web/worker/config"

	"github.com/lib/pq"
)

type DBPostgres struct {
	conn *sql.DB
}

func GetConnectionPostgres() (*DBPostgres, error) {
	cfg := config.GetConfig()
	c, err := pq.NewConnectorConfig(*cfg)
	if err != nil {
		return nil, err
	}

	// Create connection pool.
	db := sql.OpenDB(c)
	//defer db.Close()

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)

	// Make sure it works.
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DBPostgres{
		conn: db,
	}, nil
}

type Res struct {
	num int
	fib int
}

func (db *DBPostgres) Get(ctx context.Context, num int) (*Res, error) {
	con, err := db.conn.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.QueryContext(ctx, "SELECT * FROM numbers WHERE num = $1", num)
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		res := Res{}
		rows.Scan(&res)

		return &res, nil
	}

	return nil, errors.New("not found")
}

func (dbp *DBPostgres) Insert(ctx context.Context, num, calc int) error {
	con, err := dbp.conn.Conn(ctx)
	if err != nil {
		return err
	}
	defer con.Close()

	query := `INSERT INTO numbers (num, fib)
		VALUES ($1, $2)
		ON CONFLICT (num) DO NOTHING;`
	_, err = con.ExecContext(ctx, query, num, calc)
	return err
}
