package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	mssql "github.com/denisenkom/go-mssqldb"
)

var (
	debug         = flag.Bool("debug", false, "enable debugging")
	password      = flag.String("password", "", "the database password")
	port     *int = flag.Int("port", 1433, "the database port")
	server        = flag.String("server", "", "the database server")
	user          = flag.String("user", "", "the database user")
)

const (
	createTableSql      = "CREATE TABLE TestAnsiNull (bitcol bit, charcol char(1));"
	dropTableSql        = "IF OBJECT_ID('TestAnsiNull', 'U') IS NOT NULL DROP TABLE TestAnsiNull;"
	insertQuery1        = "INSERT INTO TestAnsiNull VALUES (0, NULL);"
	insertQuery2        = "INSERT INTO TestAnsiNull VALUES (1, 'a');"
	selectNullFilter    = "SELECT bitcol FROM TestAnsiNull WHERE charcol = NULL;"
	selectNotNullFilter = "SELECT bitcol FROM TestAnsiNull WHERE charcol <> NULL;"
)

func main() {
	flag.Parse()

	if *debug {
		fmt.Printf(" password:%s\n", *password)
		fmt.Printf(" port:%d\n", *port)
		fmt.Printf(" server:%s\n", *server)
		fmt.Printf(" user:%s\n", *user)
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d", *server, *user, *password, *port)
	if *debug {
		fmt.Printf(" connString:%s\n", connString)
	}

	connector, err := mssql.NewConnector(connString)
	if err != nil {
		log.Println(err)
		return
	}

	// With ANSI_NULLS set to ON, compare NULL data with = NULL or <> NULL will return 0 rows
	connector.SessionInitSQL = "SET ANSI_NULLS ON"

	db := sql.OpenDB(connector)
	defer db.Close()

	db.Exec(dropTableSql)
	_, err = db.Exec(createTableSql)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Exec(dropTableSql)

	_, err = db.Exec(insertQuery1)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = db.Exec(insertQuery2)
	if err != nil {
		log.Println(err)
		return
	}

	var bitval bool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// (*Row) Scan should return ErrNoRows
	err = db.QueryRowContext(ctx, selectNullFilter).Scan(&bitval)
	if err.Error() != "sql: no rows in result set" {
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Expects an ErrNoRows error. No error is returned")
		}
		return
	}

	// (*Row) Scan should return ErrNoRows
	err = db.QueryRowContext(ctx, selectNotNullFilter).Scan(&bitval)
	if err.Error() != "sql: no rows in result set" {
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Expects an ErrNoRows error. No error is returned")
		}
		return
	}

	// Set ANSI_NULLS to OFF
	connector.SessionInitSQL = "SET ANSI_NULLS OFF"

	// (*Row) Scan should copy data to bitval
	err = db.QueryRowContext(ctx, selectNullFilter).Scan(&bitval)
	if err != nil {
		log.Println(err)
		return
	}
	if bitval != false {
		log.Println("Incorrect value retrieved.")
		return
	}

	// (*Row) Scan should copy data to bitval
	err = db.QueryRowContext(ctx, selectNotNullFilter).Scan(&bitval)
	if err != nil {
		log.Println(err)
		return
	}
	if bitval != true {
		log.Println("Incorrect value retrieved.")
		return
	}
}
