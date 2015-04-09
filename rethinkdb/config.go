package rethinkdb

type Config struct {
	Address  string
	MaxIdle  int
	MaxOpen  int
	Database string
}
