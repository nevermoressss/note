package orm

import (
	"database/sql"
	"testing"
)

func TestConnect(t *testing.T) {
	type args struct {
		dsn string
	}
	tests := []struct {
		name    string
		args    args
		want    *sql.DB
		wantErr bool
	}{
		{
			name: "testConnect",
			args: args{
				dsn: `root:123456@(10.0.1.78:3306)/orm_db?charset=utf8mb4&parseTime=True&loc=PRC`,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Connect(tt.args.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = got.Close()
		})
	}
}

func TestTable(t *testing.T) {

	got, err := Connect(`root:123456@(10.0.1.78:3306)/orm_db?charset=utf8mb4&parseTime=True&loc=PRC`)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	type args struct {
		db        *sql.DB
		tableName string
	}
	tests := []struct {
		name string
		args args
		want func() *Query
	}{
		{
			name: "true test",
			args: args{
				db:        got,
				tableName: "user",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//users:=Table(tt.args.db, tt.args.tableName)
			//users().
		})
	}
}

func TestSelect(t *testing.T) {
	//全局变量ormDB和users
	ormDB, _ := Connect("root:123456@(10.0.1.78:3306)/orm_db?charset=utf8mb4&parseTime=True&loc=PRC")
	defer ormDB.Close()

	users := Table(ormDB, "user")
	//调用
	//users().Insert(...)
	var user User
	err := users().Where("first_name = 'Tom'", map[string]interface{}{
		"id": []int{1, 2, 3, 4},
	}).WhereNot(&User{LastName: "Cat"}).Only("last_name").Select(&user)
	if err != nil {
		t.Fatal(err)
	}
	var userMore []User
	err = users().Where("first_name = 'Tom'").Order("id desc").Select(&userMore)
	if err != nil {
		t.Fatal(err)
	}
	var userMoreP []*User
	err = users().Where("first_name = 'Tom'").Select(&userMoreP)
	if err != nil {
		t.Fatal(err)
	}
	var lastName string
	err = users().Where(&User{FirstName: "Tom"}).Only("last_name").Select(&lastName)
	if err != nil {
		t.Fatal(err)
	}
	var lastNames []string
	err = users().Where(map[string]interface{}{
		"first_name": "Tom",
	}).Only("last_name").Select(&lastNames)
	if err != nil {
		t.Fatal(err)
	}
	var userM map[string]interface{}
	err = users().Where(&User{FirstName: "Tom"}).Only("last_name").Select(&userM)
	if err != nil {
		t.Fatal(err)
	}
	var userMS []map[string]interface{}
	err = users().Where("age > 10").Only("last_name", "age").Limit(100).Select(&userMS)
	if err != nil {
		t.Fatal(err)
	}
}
