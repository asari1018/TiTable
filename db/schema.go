package db

// schema.go provides data models in DB
import (
	"time"
)

// Task corresponds to a row in `tasks` table
type Task struct {
	ID         uint64    `db:"id"`
	Title      string    `db:"title"`
	Class      string    `db:"class"`
	CreatedAt  time.Time `db:"created_at"`
	DeadlineAt time.Time `db:"deadline_at"`
	IsDone     bool      `db:"is_done"`
	TaskLevel  uint64    `db:"task_level"`
}

//User corresponds to a row in 'users' table
type User struct {
	ID       uint64    `db:"id"`
	Name     string    `db:"name"`
	EmailID  string    `db:"email_id"`
	UserAuth string    `db:"user_auth"`
	Password string    `db:"password"`
	LastTime time.Time `db:"last_time"`
}

//Class corresponds to a row in 'classes' table
type Class struct {
	ID      uint64    `db:"id"`
	Class   string    `db:"class"`
	UID     uint64    `db:"uid"`
	Comment string    `db:"comment"`
	Start   time.Time `db:"start"`
	End     time.Time `db:"end"`
	URL     string    `db:"url"`
	X       uint64    `db:"x"`
	Y       uint64    `db:"y"`
	Length  uint64    `db:"length"`
}

//UserInfo corresponds to a row in 'user_info' table
type UserInfo struct {
	Task_ID uint64 `db:"task_id"`
	User_ID uint64 `db:"user_id"`
}
