package reply

type ProtocolErrReply struct {
	Msg string
}

func (r *ProtocolErrReply)ToBytes()[]byte{
	return []byte("-ERR Protocol error: '" + r.Msg + "'\r\n")
}