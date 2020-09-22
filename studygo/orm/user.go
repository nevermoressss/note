package orm

import "time"

type User struct {
	ID        int64     `json:"id"`         // 自增主键
	Age       int64     `json:"age"`        // 年龄
	FirstName string    `json:"first_name"` // 姓
	LastName  string    `json:"last_name"`  // 名
	Email     string    `json:"email"`      // 邮箱地址
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}


func NewUser(ID int64, age int64, firstName string, lastName string, email string, createdAt time.Time, updatedAt time.Time) *User {
	return &User{ID: ID, Age: age, FirstName: firstName, LastName: lastName, Email: email, CreatedAt: createdAt, UpdatedAt: updatedAt}
}
