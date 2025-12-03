package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cunkz/goyummy/bin/config"
)

/* ===============================
   INIT DBs: POSTGRES + MYSQL + MONGO
================================ */

var (
	PostgresDBs = make(map[string]*sql.DB)
	MySQLDBs    = make(map[string]*sql.DB)
	MongoDBs    = make(map[string]*mongo.Database)
)

func InitDatabases(cfg *config.AppConfig) error {
	ctx := context.Background()

	for _, db := range cfg.Databases {
		switch db.Engine {

		// -----------------------------------------------------
		// POSTGRES + MYSQL (same pooling behavior)
		// -----------------------------------------------------
		case "postgres", "mysql":
			conn, err := sql.Open(db.Engine, db.URI)
			if err != nil {
				return fmt.Errorf("%s (%s) error: %v", db.Engine, db.Name, err)
			}

			// Set Pooling
			if db.Pool.Max > 0 {
				conn.SetMaxOpenConns(db.Pool.Max)
			}
			if db.Pool.Min > 0 {
				conn.SetMaxIdleConns(db.Pool.Min)
			}
			conn.SetConnMaxLifetime(5 * time.Minute)

			// Store based on engine
			if db.Engine == "postgres" {
				PostgresDBs[db.Name] = conn
			} else {
				MySQLDBs[db.Name] = conn
			}

		// -----------------------------------------------------
		// MONGO DB
		// -----------------------------------------------------
		case "mongo":
			opts := options.Client().ApplyURI(db.URI)

			// POOLING
			if db.Pool.Max > 0 {
				opts.SetMaxPoolSize(uint64(db.Pool.Max))
			}
			if db.Pool.Min > 0 {
				opts.SetMinPoolSize(uint64(db.Pool.Min))
			}

			client, err := mongo.Connect(ctx, opts)
			if err != nil {
				return fmt.Errorf("mongo (%s) error: %v", db.Name, err)
			}

			MongoDBs[db.Name] = client.Database(db.Name)
		}
	}

	return nil
}
