package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/abetd/sqlgen/examples"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable := `
	create table users (id integer not null primary key, name text);
	create table items (id integer not null primary key, name text, kana text);
	insert into users (id, name) values (1, 'test1'), (2, 'test2');
	insert into items (id, name, kana) values (1, 'test1', 'TEST1'), (2, 'test2', 'TEST2');
	`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Printf("%q: %s\n", err, createTable)
		return
	}

	if err := selectUsers(db); err != nil {
		log.Fatal(err)
	}

	if err := selectItems(db); err != nil {
		log.Fatal(err)
	}
}

func selectUsers(db *sql.DB) error {
	e := examples.SelectUsersQueryElem{
		ID:    2,
		Name:  "test3",
		Names: []interface{}{"test1", "test2"},
	}
	q, err := e.Query()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query(q.SQL, q.Args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		log.Printf("id: %d, name: %s\n", id, name)
	}
	return nil
}

func selectItems(db *sql.DB) error {
	e := examples.SelectItemsQueryElem{
		ID:                 1,
		IsSelectMultiNames: true,
		Where:              "(name LIKE ? OR kana LIKE ?)",
		Sep:                "AND",
		Names:              []interface{}{"%test%", "%TEST%"},
	}
	q, err := e.Query()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query(q.SQL, q.Args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		var kana string
		if err := rows.Scan(&id, &name, &kana); err != nil {
			return err
		}
		log.Printf("id: %d, name: %s, kana: %s\n", id, name, kana)
	}
	return nil
}
