package seqno

func MigrateSQL() string {
	return `
CREATE TABLE seqno_generator (
id bigint(18) auto_increment primary key,
logic_id CHAR(32),
prefix CHAR(32),
last_id bigint(18),
created_at datetime,
updated_at datetime,
deleted_at datetime
)
`
}
