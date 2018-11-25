package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var (
	db *sql.DB
)

type Item struct {
	Id        int64
	Name      string
	SellPrice int64
	JSON      string
}

func Open() {
	var err error

	db, err = sql.Open("mysql", "wow:wowpassword@tcp(127.0.0.1:3306)/wow")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Close() {
	db.Close()
}

func SaveItem(item Item) {
	sqlString := "INSERT IGNORE INTO items ( id, name, sellPrice, json ) VALUES ( ?, ?, ?, ? )"

	_, err := db.Exec(sqlString, item.Id, item.Name, item.SellPrice, item.JSON)
	if err != nil {
		fmt.Println("dbSaveItem Exec:", err)
	}
}

func LookupItem(id int64) (Item, bool) {
	var item Item

	sqlString := "SELECT * FROM items WHERE id = " + fmt.Sprintf("%d", id) + " LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&item.Id, &item.Name, &item.SellPrice, &item.JSON)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("dbLookupItem Scan:", err)
		}
		return item, false
	}

	return item, true
}
