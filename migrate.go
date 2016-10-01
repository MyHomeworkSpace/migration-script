package main

import (
	"log"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Column struct {
    field string
    colType string
    nullable string
    key string
    defaultVal interface{}
    extra string
}

var DB *sql.DB

func InitDatabase() {
	var err error

	DB, err = sql.Open("mysql", config.Database.Username+":"+config.Database.Password + "@(" + config.Database.Host + ")/" + config.OldDB)
	if err != nil {
		panic(err)
	}
	err = DB.Ping()
	if err != nil {
		panic(err)
	}
}

func MigrateExactly(oldName string, newName string) {
    log.Printf("Migrating table '%s' to '%s'...\n", oldName, newName)
    createNewStmt, err := DB.Prepare("CREATE TABLE `" + config.NewDB + "`." + newName + " LIKE `" + config.OldDB + "`." + oldName)
    if err != nil {
    	log.Fatal(err)
    }
    _, err = createNewStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }

    insertNewStmt, err := DB.Prepare("INSERT `" + config.NewDB + "`." + newName + " SELECT * FROM `" + config.OldDB + "`." + oldName)
    if err != nil {
    	log.Fatal(err)
    }
    _, err = insertNewStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }
}

func main() {
	log.Println("Migration script")

	InitConfig()
	InitDatabase()

	// announcements
    MigrateExactly("planner_announcements", "announcements")
    log.Println("Migrated announcements!")

	// feedback
    log.Printf("Migrating table 'feedback' to 'feedback'...\n")
    feedbackCreateStmt, err := DB.Prepare("CREATE TABLE `" + config.NewDB + "`.`feedback`(`id` int(11) NOT NULL AUTO_INCREMENT, `userId` int(11) DEFAULT NULL, `type` varchar(10) DEFAULT NULL, `text` text, `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=latin1;")
    if err != nil {
    	log.Fatal(err)
    }
    _, err = feedbackCreateStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }

    feedbackInsertStmt, err := DB.Prepare("INSERT `" + config.NewDB + "`.feedback SELECT `feedbackId` AS `id`, " + config.OldDB + ".users.id AS `userId`, " + config.OldDB + ".feedback.`type`, `msg` AS `text`, `timestamp` FROM " + config.OldDB + ".feedback INNER JOIN " + config.OldDB + ".users ON " + config.OldDB + ".feedback.username = " + config.OldDB + ".users.username")
    if err != nil {
    	log.Fatal(err)
    }
    _, err = feedbackInsertStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }
    log.Println("Migrated feedback!")

	// fridays
    MigrateExactly("planner_fridays", "fridays")
    fridaysIDStmt, err := DB.Prepare("ALTER TABLE `" + config.NewDB + "`.fridays CHANGE entryId id INT")
    if err != nil {
    	log.Fatal(err)
    }
    _, err = fridaysIDStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }
    log.Println("Migrated fridays!")

	// users
    MigrateExactly("users", "users")
	usersColStmt, err := DB.Prepare("ALTER TABLE `" + config.NewDB + "`.users ADD COLUMN showMigrateMessage INT AFTER canAnnouncements")
	if err != nil {
		log.Fatal(err)
	}
	_, err = usersColStmt.Exec()
	if err != nil {
		log.Fatal(err)
	}
	usersFlagStmt, err := DB.Prepare("UPDATE `" + config.NewDB + "`.users SET showMigrateMessage=1")
	if err != nil {
		log.Fatal(err)
	}
	_, err = usersFlagStmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

    log.Println("Migrated users!")

    // classes and homework

}
