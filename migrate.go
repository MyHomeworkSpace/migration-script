package main

import (
	"log"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

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

    // classes
    log.Printf("Migrating table 'planner_sections' to 'classes'...\n")
	classesCreateStmt, err := DB.Prepare("CREATE TABLE `" + config.NewDB + "`.`classes` (`id` int(11) NOT NULL AUTO_INCREMENT, `name` text, `teacher` text, `userId` int(11) DEFAULT NULL, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=latin1")
	if err != nil {
    	log.Fatal(err)
    }
    _, err = classesCreateStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }
	classesInsertStmt, err := DB.Prepare("INSERT `" + config.NewDB + "`.classes SELECT sectionGid AS `id`, `name`, \"\" AS `teacher`, userId FROM " + config.OldDB + ".planner_sections")
    if err != nil {
    	log.Fatal(err)
    }
    _, err = classesInsertStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }
    log.Printf("Migrated classes!\n")

	// homework
    log.Printf("Migrating table 'planner_events' to 'homework'...\n")
	homeworkCreateStmt, err := DB.Prepare("CREATE TABLE `" + config.NewDB + "`.`homework` (`id` int(11) NOT NULL AUTO_INCREMENT, `name` text, `due` date DEFAULT NULL, `desc` text, `complete` varchar(45) DEFAULT NULL, `classId` int(11) DEFAULT NULL, `userId` int(11) DEFAULT NULL, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=latin1")
	if err != nil {
    	log.Fatal(err)
    }
    _, err = homeworkCreateStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }
	homeworkInsertStmt, err := DB.Prepare("INSERT `" + config.NewDB + "`.homework SELECT eventId AS `id`, `text` AS `name`, `date` AS `due`, \"\" AS `desc`, `done` AS `complete`, `" + config.OldDB + "`.planner_sections.sectionGid AS `classId`, `" + config.OldDB + "`.planner_events.`userId` FROM `" + config.OldDB + "`.planner_events INNER JOIN `" + config.OldDB + "`.planner_sections ON ((`" + config.OldDB + "`.planner_sections.userId = `" + config.OldDB + "`.planner_events.userId) AND (`" + config.OldDB + "`.planner_sections.sectionIndex = `" + config.OldDB + "`.planner_events.sectionIndex)) WHERE `text` != \"\" AND `text` != \"none\"")
    if err != nil {
    	log.Fatal(err)
    }
    _, err = homeworkInsertStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }
    log.Printf("Migrated homework!\n")

	// sessions
	log.Printf("Creating table 'sessions'...\n")
	sessionsCreateStmt, err := DB.Prepare("CREATE TABLE `" + config.NewDB + "`.`sessions` (`id` varchar(255) NOT NULL, `userId` int(11) DEFAULT NULL, `username` text, `timestamp` TIMESTAMP, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8")
	if err != nil {
    	log.Fatal(err)
    }
    _, err = sessionsCreateStmt.Exec()
    if err != nil {
    	log.Fatal(err)
    }
    log.Printf("Created sessions!\n")
}
