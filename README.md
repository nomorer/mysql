Package mysql provides a multi sql.DB manager and map query data to container.
Getting Started
=============
    package main
    
    import (
    	"fmt"
    
    	"github.com/just-go/mysql"
    )
    
    func main() {
    
    	err := mysql.RegisterDatabase("test", "root:root@/test")
    	if err != nil {
    		panic(err)
    	}
    
    	type User struct {
    		Id    int64  `db:"id"`
    		Name  string `db:"name"`
    		Email string `db:"email"`
    	}
    	var user User
    	err = mysql.GetDB("test").QueryRow(&user, "select * from users")
    	if err != nil {
    		panic(err)
    	}
    	fmt.Println(user)
    
    	var users []User
    	err = mysql.GetDB("test").QueryRows(&users, "select * from users")
    	if err != nil {
    		panic(err)
    	}
    	fmt.Println(users)
    }
