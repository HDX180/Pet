package DeviceManageServer

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func OpenDB(dsn string) {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Panic(err)
		return
	}
}

func CloseDB() {
	db.Close()
}

//DEVINFO--->| ID | codeID |

func db_getDevInfo(m *map[int]*struDevInfo) int {
	var num int
	// err := db.QueryRow("SELECT COUNT(*) FROM DEVINFO").Scan(&num)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	rows, err := db.Query("SELECT * FROM DEVINFO")
	for rows.Next() {
		devinfo := new(struDevInfo)
		err = rows.Scan(&devinfo.index, &devinfo.codeID)
		if err != nil {
			log.Fatal(err)
		}
		devinfo.status = false //初始化为不在线的状态
		(*m)[devinfo.codeID] = devinfo
		num++
	}
	return num
}

func db_getDevIndex(codeID int) int {
	index := 0
	err := db.QueryRow("SELECT ID FROM DEVINFO WHERE CODEID = ?", codeID).Scan(&index)
	if err != nil {
		Logger.Debug(err.Error())
	}
	return index
}
