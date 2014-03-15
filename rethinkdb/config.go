package rethinkdb

type Config struct {
	Address   string
	MaxIdle   int
	MaxActive int
	Database  string
}
