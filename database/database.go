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

type Auction struct {
	Auc           int64
	Item          int64
	Owner         string
	Bid           int64
	Buyout        int64
	Quantity      int64
	TimeLeft      string
	Rand          int64
	Seed          int64
	Context       int64
	HasBonusLists bool
	HasModifiers  bool
	PetBreedId    int64
	PetLevel      int64
	PetQualityId  int64
	PetSpeciesId  int64
	JSON          string
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
		fmt.Println("SaveItem Exec:", err, item)
	}
}

func LookupItem(id int64) (Item, bool) {
	var item Item

	sqlString := "SELECT * FROM items WHERE id = " + fmt.Sprintf("%d", id) + " LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&item.Id, &item.Name, &item.SellPrice, &item.JSON)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("LookupItem Scan:", err, item)
		}
		return item, false
	}

	return item, true
}

func SaveAuction(auction Auction) {
	var sqlString string
	var err error

	// If the row already exists we want to update it.
	if current, ok := LookupAuction(auction.Auc); ok {
		if auction != current {
			sqlString = "UPDATE auctions SET item = ?, owner = ?, bid = ?, buyout = ?, quantity = ?, timeLeft = ?, rand = ?, seed = ?, context = ?, hasBonusLists = ?, hasModifiers = ?, petBreedId = ?, petLevel = ?, petQualityId = ?, petSpeciesId = ?, json = ? WHERE auc = ?"
			_, err = db.Exec(sqlString, auction.Item, auction.Owner, auction.Bid, auction.Buyout, auction.Quantity, auction.TimeLeft, auction.Rand, auction.Seed, auction.Context, auction.HasBonusLists, auction.HasModifiers, auction.PetBreedId, auction.PetLevel, auction.PetQualityId, auction.PetSpeciesId, auction.JSON, auction.Auc)
			if err != nil {
				fmt.Println("SaveAuction Exec(UPDATE):", err, auction)
			}
		}
	} else {
		sqlString = "INSERT INTO auctions ( auc, item, owner, bid, buyout, quantity, timeLeft, rand, seed, context, hasBonusLists, hasModifiers, PetBreedId, petLevel, petQualityId, petSpeciesId, json ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"
		_, err = db.Exec(sqlString, auction.Auc, auction.Item, auction.Owner, auction.Bid, auction.Buyout, auction.Quantity, auction.TimeLeft, auction.Rand, auction.Seed, auction.Context, auction.HasBonusLists, auction.HasModifiers, auction.PetBreedId, auction.PetLevel, auction.PetQualityId, auction.PetSpeciesId, auction.JSON)
		if err != nil {
			fmt.Println("SaveAuction Exec(INSERT):", err, auction)
		}
	}

}

func LookupAuction(auc int64) (Auction, bool) {
	var auction Auction
	var lastUpdated string

	sqlString := "SELECT * FROM auctions WHERE auc = " + fmt.Sprintf("%d", auc) + " LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&auction.Auc, &auction.Item, &auction.Owner, &auction.Bid, &auction.Buyout, &auction.Quantity, &auction.TimeLeft, &auction.Rand, &auction.Seed, &auction.Context, &auction.HasBonusLists, &auction.HasModifiers, &auction.PetBreedId, &auction.PetLevel, &auction.PetQualityId, &auction.PetSpeciesId, &auction.JSON, &lastUpdated)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("LookupAuction Scan:", err)
		}
		return auction, false
	}

	return auction, true
}

func countRows(table string) int64 {
	var count int64

	sqlString := "SELECT count(*) FROM " + table
	rows := db.QueryRow(sqlString)
	err := rows.Scan(&count)
	if err != nil {
		fmt.Println("countRows Scan:", err)
		return -1
	}

	return count
}

func CountItems() int64 {
	return countRows("items")
}

func CountAuctions() int64 {
	return countRows("auctions")
}
