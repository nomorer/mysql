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
    
    	//table
    	type User struct {
    		Id    int64  `db:"id"`
    		Name  string `db:"name"`
    		Email string `db:"email"`
    	}
    
    	var id int64
    	if err = mysql.GetDB("test").QueryRow(&id, "select id from users limit 1"); err != nil {
    		panic(err)
    	}
    	fmt.Println(id)
    
    	var name string
    	if err = mysql.GetDB("test").QueryRow(&name, "select name from users limit 1"); err != nil {
    		panic(err)
    	}
    	fmt.Println(name)
    
    	var names []string
    	if err = mysql.GetDB("test").QueryRows(&names, "select name from users"); err != nil {
    		panic(err)
    	}
    	fmt.Println(names)
    
    	var emails []*string
    	if err = mysql.GetDB("test").QueryRows(&emails, "select name from users"); err != nil {
    		panic(err)
    	}
    	fmt.Println(emails)
    
    	var user User
    	err = mysql.GetDB("test").QueryRow(&user, "select id, name from users")
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
    
    	var users2 []*User
    	err = mysql.GetDB("test").QueryRows(&users2, "select * from users")
    	if err != nil {
    		panic(err)
    	}
    	fmt.Println(users2)
    }
