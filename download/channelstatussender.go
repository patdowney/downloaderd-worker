package download

type ChannelStatusSender struct {
	UpdatesSent   uint
	StatusChannel chan StatusUpdate
}

func (s *ChannelStatusSender) SendUpdate(update StatusUpdate) {
	s.StatusChannel <- update
	s.UpdatesSent += 1
}
