package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var dbError error

func databaseInit() {
	db, dbError = sql.Open("mysql", ":@/test")

	if dbError != nil {
		fmt.Printf("Database: Conn Error")
		db.Close()
		return
	}
}

func databaseClose() {
	db.Close()
}

func databaseSignIn(username string, password string) string {
	rows, e := db.Query("select count(*) from user_record where username = ? and password = ?", username, password)

	if e != nil {
		fmt.Println("Database: Query Error.", e.Error())
		return "failed"
	}

	count := 0
	if rows.Next() {
		e = rows.Scan(&count)
	}
	if e != nil {
		fmt.Println("Database: Scan Error.", e.Error())
		return "failed"
	}

	if count > 0 {
		return "success"
	} else {
		return "failed"
	}
}

func databaseSignUp(username string, password string, email string) string {
	rows, e := db.Query("select count(*) from user_record where username = ?", username)

	if e != nil {
		fmt.Println("Database: Query Error.")
		return "failed"
	}

	count := 0
	if rows.Next() {
		e = rows.Scan(&count)
	}
	if e != nil {
		fmt.Println("Database: Scan Error.")
		return "failed"
	}
	if count > 0 {
		return "failed"
	}

	_, e = db.Exec("insert into user_record set username = ?, password = ?, email = ?", username, password, email)
	if e != nil {
		fmt.Println("Database: Execute Error.")
		return "failed"
	}

	return "success"
}

func main() {
	databaseInit()

	result := databaseSignIn("kb24", "nddndd")
	fmt.Println(result)
	result = databaseSignUp("chenzhuokb2", "nddndd", "kb24@cmu.edu")
	fmt.Println(result)

	/*rows, e := db.Query("select username from auth_user")
	if e != nil {
		fmt.Println("Database: Query Error.")
		db.Close()
		return
	}

	i := 0
	for rows.Next() {
		i++
		var username string
		e := rows.Scan(&username)

		if e != nil {
			fmt.Println("Database: Scan Error")
			db.Close()
		} else {
			fmt.Println("Username: ", username)
		}
	}*/

	databaseClose()

}
