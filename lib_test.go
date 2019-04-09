package seqno

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
	"testing"
	"time"
)

func TestNext(t *testing.T) {
	db, err := gorm.Open("mysql", "root:root@tcp(mysql:3306)/bindo?parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(db.Debug().Exec(MigrateSQL()).Error)
	seqno := NewSeqNoGenerator(db, "test-logic-id").PrefixFormat("2006-01-02").Locker(NewRedisLocker(
		&redis.Pool{
			MaxIdle:     5,
			IdleTimeout: 30 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", "redis:6379")
			},
		},
	)).Step(2)
	seqno.Next()
	seqno.Next()
	num, err := seqno.Next()
	if err != nil {
		t.Fatal(err)
	}
	t.Error(num)
}

func BenchmarkNext(b *testing.B) {
	db, err := gorm.Open("mysql", "root:root@tcp(mysql:3306)/xxx?parseTime=true")
	if err != nil {
		b.Fatal(err)
	}
	b.Log(db.Debug().Exec(MigrateSQL()).Error)
	seqno := NewSeqNoGenerator(db, "test-logic-id").PrefixFormat("2006-01-02").Locker(NewRedisLocker(
		&redis.Pool{
			MaxIdle:     5,
			IdleTimeout: 30 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", "redis:6379")
			},
		},
	)).Step(1)

	for n := 0; n < b.N; n++ {
		num,err := seqno.Next()
		if err != nil {
			b.Fatal(err)
		}
		fmt.Println(num)
	}
}
