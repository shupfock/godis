package redis

type Connection interface {
	Write([]byte) error

	SubsChannel(channel string)
	UnSubsChannel(channel string)
	SubsCount() int
	GetChannels() []string
}
