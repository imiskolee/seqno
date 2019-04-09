# SeqNo Generator

a distributed sequence number generator based on MySQL.

# Why?

Need a human call number on some cases(Restaurant etc...).


# MySQL schema

```sql
CREATE TABLE `seqno_generator` (
  `id` bigint(18) NOT NULL AUTO_INCREMENT,
  `logic_id` char(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `prefix` char(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `last_id` bigint(18) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni` (`logic_id`,`prefix`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```

# Examples

```go
    
    //Open a gorm instance
    db, err := gorm.Open("mysql", "root:root@tcp(mysql:3306)/xxx?parseTime=true")
	if err != nil {
		b.Fatal(err)
	}
    
	seqno := NewSeqNoGenerator(db, "test-logic-id").
		PrefixFormat("2006-01-02").
		Step(1).
		Locker(NewRedisLocker(
		&redis.Pool{
			MaxIdle:     5,
			IdleTimeout: 30 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", "redis:6379")
			},
		},
	))
	num,err := seqno.Next()
```

## Options

### Logic ID

`Logic ID` represents a concrete business logic(like `order`,`user`).

### Step

Incremental step. default is  1

### StartWith

Starting number. default is 1

### PrefixFormat

date time format. default is empty. it's mean always use same increment queue. 
if set `PrefixFormat` to `2006-01-02`, it's mean go to zero once a day.

### Locker

locker is a distributed locker interface. already have redis & ETCD implementation on standard library.

## Technical Model

1. Lock currently logic id.
2. Insert or Update New Value to MySQL(depend  feature `DUPLICATE KEY UPDATE` ).
3. get last id.
4. Unlock logic id.

2 locker call and  2 MySQL call on a full flow.


**ITS ON DEVELOPING STAGE,DONT USE IT ON PRODUCTION.**

