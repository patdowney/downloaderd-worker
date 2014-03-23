package download

type StatusSender interface {
	SendUpdate(StatusUpdate)
}
