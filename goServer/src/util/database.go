package util

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS user_score (id int(11) NOT NULL AUTO_INCREMENT,
	username VARCHAR(50) NOT NULL, score INT NOT NULL, 
	PRIMARY KEY (id), UNIQUE KEY username (username))`)
	if err != nil {
		fmt.Println("Failed to create build-in user_score table")
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

func DatabaseCreateDBFile() error {
	fmt.Println("Database: createDBFile")

	var e error

	if checkFileExist("/tmp/users.csv") {
		e = os.Remove("/tmp/users.csv")
	}

	if e != nil {
		fmt.Println("Database: Error.", e.Error())
		return e
	}

	_, e = db.Exec("select username, password, email from user_record into outfile '/tmp/users.csv' fields terminated by ',' enclosed by '\"' lines terminated by '\n'")
	if e != nil {
		fmt.Println("Database: Execute Error.", e.Error())
	}
	return e
}

func DatabaseLoadDBFile() error {
	fmt.Println("loadDBFile")

	_, e := db.Exec("load data infile '/tmp/output.csv' ignore into table user_record fields terminated by ',' enclosed by '\"' lines terminated by '\n' (username, password, email)")
	if e != nil {
		fmt.Println("Database: Execute Error.", e.Error())
	}
	return e
}

func checkFileExist(filename string) bool {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}

	return true
}

func DatabaseReadScore(username string) (int, error) {
	rows, e := db.Query("select score from user_score where username = ?", username)

	if e != nil {
		fmt.Println("Database: Query Error.", e.Error())
		return -1, e
	}

	score := -1
	if rows.Next() {
		e = rows.Scan(&score)
	}
	if e != nil {
		fmt.Println("Database: Scan Error.", e.Error())
		return -1, e
	}

	return score, nil
}

func DatabaseUpdateScore(username string, score int) error {
	rows, e := db.Query("select count(*) from user_score where username = ?", username)
	if e != nil {
		fmt.Println("Database: Query Error.", e.Error())
		return e
	}

	count := 0
	if rows.Next() {
		e = rows.Scan(&count)
	}
	if e != nil {
		fmt.Println("Database: Scan Error.", e.Error())
		return e
	}

	if count > 0 {
		_, e = db.Exec("update user_score set score = ? where username = ?", score, username)
		if e != nil {
			fmt.Println("Database: Execute Error.", e.Error())
			return e
		}

	} else {
		_, e = db.Exec("insert into user_score set username = ?, score = ?", username, score)
		if e != nil {
			fmt.Println("Database: Execute Error.", e.Error())
			return e
		}
	}

	return nil
}
