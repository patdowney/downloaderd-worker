package download

import (
	"time"
)

type StatusUpdate struct {
	OrderId   string
	BytesRead uint64
	Checksum  string
	Time      time.Time
	Finished  bool
}

func NewStatusUpdate(orderId string, bytesRead uint64, checksum string, updateTime time.Time) *StatusUpdate {
	return &StatusUpdate{
		OrderId:   orderId,
		BytesRead: bytesRead,
		Checksum:  checksum,
		Time:      updateTime,
		Finished:  false}
}

func NewFinishedStatusUpdate(orderId string, bytesRead uint64, checksum string, updateTime time.Time) *StatusUpdate {
	return &StatusUpdate{
		OrderId:   orderId,
		BytesRead: bytesRead,
		Checksum:  checksum,
		Time:      updateTime,
		Finished:  true}
}
