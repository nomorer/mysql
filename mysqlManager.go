package mysql

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlManager struct {
	dbs  map[string]*Mysql
	lock *sync.Mutex
}

const (
	driverName string = "mysql"
)

var (
	mysqlManager *MysqlManager
)

func init() {
	mysqlManager = &MysqlManager{
		dbs:  make(map[string]*Mysql),
		lock: new(sync.Mutex),
	}
}

//RegisterDB register *sql.DB
func RegisterDB(aliasName string, db *sql.DB) {
	mysqlManager.lock.Lock()
	defer mysqlManager.lock.Unlock()

	mysqlManager.dbs[aliasName] = NewMysql(db)
}

//RegisterDatabase opens a database specified by mysql driver and its data source name
//using params, you can set the database conns
func RegisterDatabase(aliasName, dataSource string, params ...int) error {
	db, err := sql.Open(driverName, dataSource)
	if err != nil {
		return err
	}

	for i, v := range params {
		switch i {
		case 0:
			db.SetMaxIdleConns(v)
		case 1:
			db.SetMaxOpenConns(v)
		}
	}

	RegisterDB(aliasName, db)
	return nil
}

// GetDB Get *sql.DB from registered database by db alias name.
func GetDB(aliasName string) *Mysql {
	mysqlManager.lock.Lock()
	defer mysqlManager.lock.Unlock()

	if mysql, ok := mysqlManager.dbs[aliasName]; ok {
		return mysql
	}

	panic(fmt.Sprintf("%s Not Found", aliasName))
}
