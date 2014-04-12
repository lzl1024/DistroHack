package util

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var dbError error

func DatabaseInit(isSN bool) {
	db, dbError = sql.Open("mysql", ":@/test")

	if dbError != nil {
		fmt.Printf("Database: Conn Error")
		db.Close()
		return
	}
	
	// create user_record if not exists
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS user_record (id int(11) NOT NULL AUTO_INCREMENT,
	username VARCHAR(50) NOT NULL, password VARCHAR(50) NOT NULL, email VARCHAR(50) NOT NULL, 
	PRIMARY KEY (id), UNIQUE KEY username (username))`)
	if err != nil {
		fmt.Println("Failed to create build-in user_record table")
	}
	
	// if is SN add admin into database
	if isSN {
		_, err := db.Exec(`insert ignore into user_record (username, password, email) 
		values ('admin','admin', 'admin@admin'); `)
		if err != nil {
			fmt.Println("Failed to insert build-in admin")
		}
	}
	
}

func DatabaseClose() {
	db.Close()
}

func DatabaseSignIn(username string, password string) string {
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


func DatabaseSignUp(username string, password string, email string) string {
	/*rows, e := db.Query("select count(*) from user_record where username = ? or email = ?", username, email)

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
		return "This email/username has already been registered!"
	}*/

	_, e := db.Exec("insert into user_record set username = ?, password = ?, email = ?", username, password, email)
	if e != nil {
		fmt.Println("Database: Execute Error.", e.Error())
		return "failed"
	}

	return "success"
}


func DatabaseCheckUser(username string, email string) string {
	rows, e := db.Query("select count(*) from user_record where username = ? or email = ?", username, email)

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
		return "This email/username has already been registered!"
	} else {
		return "success"
	}
}