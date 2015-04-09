package download

// ChannelStatusSender ...
type ChannelStatusSender struct {
	UpdatesSent   uint
	StatusChannel chan StatusUpdate
}

// SendUpdate ...
func (s *ChannelStatusSender) SendUpdate(update StatusUpdate) {
	s.StatusChannel <- update
	s.UpdatesSent++
}
