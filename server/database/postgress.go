package database

import (
	"context"
	"database/sql"
	"errors"
	"local/complex-web/server/config"

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

	dbPostgres := &DBPostgres{
		conn: db,
	}

	if err = dbPostgres.initDB(); err != nil {
		return nil, err
	}

	return dbPostgres, nil
}

type Res struct {
	Id  int
	Num int
	Fib int
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
		err := rows.Scan(&res.Id, &res.Num, &res.Fib)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		return &res, nil
	}

	return nil, errors.New("not found")
}

func (dbp *DBPostgres) initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS numbers (
		id serial PRIMARY KEY,
		num integer NOT NULL UNIQUE,
		fib integer NOT NULL
	);
    `

	_, err := dbp.conn.Exec(query)
	return err
}
