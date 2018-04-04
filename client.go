package main

import (
    "context"
    "database/sql"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/cockroachdb/cockroach-go/crdb"
)

var (
    // FQDN of cockroach db instance
    cockroachdb_fqdn string
)

func transferFunds(tx *sql.Tx, from int, to int, amount int) error {
    // Read the balance.
    var fromBalance int
    if err := tx.QueryRow(
        "SELECT balance FROM accounts WHERE id = $1", from).Scan(&fromBalance); err != nil {
        return err
    }

    if fromBalance < amount {
        return fmt.Errorf("insufficient funds")
    }

    // Perform the transfer.
    if _, err := tx.Exec(
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from); err != nil {
        return err
    }
    if _, err := tx.Exec(
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to); err != nil {
        return err
    }
    return nil
}

func init() {
    flag.StringVar(&cockroachdb_fqdn, "cockroachdb", "pg.cockroachdb.l4lb.thisdcos.directory", "The FQDN or IP address and port of cockroachdb.")
        flag.Usage = func() {
        fmt.Printf("Usage: %s [args]\n\n", os.Args[0])
        fmt.Println("Arguments:")
        flag.PrintDefaults()
    }
    flag.Parse()
}

func main() {
    db, err := sql.Open("postgres", "postgresql://root@pg.cockroachdb.l4lb.thisdcos.directory:26257/bank?sslmode=disable")
    if err != nil {
        log.Fatal("error connecting to the database: ", err)
    }

    // Create the "accounts" table.
    if _, err := db.Exec(
        "CREATE TABLE IF NOT EXISTS accounts (id INT PRIMARY KEY, balance INT)"); err != nil {
        log.Fatal(err)
    }

    // Run a transfer in a transaction.
    err = crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
        return transferFunds(tx, 1 /* from acct# */, 2 /* to acct# */, 100 /* amount */)
    })
    if err == nil {
        fmt.Println("Success")
    } else {
        log.Fatal("error: ", err)
    }
}
