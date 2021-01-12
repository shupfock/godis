package server

import (
	"godis/src/lib/sync/atomic"
	"godis/src/lib/sync/wait"
	"net"
	"time"
)

// Client 接受请求的客户端
type Client struct {
	// 与客户端的 tcp 链接
	conn net.Conn
	/*
	 * 带有timeout功能的WaitGroup, 用于优雅关闭
	 * 当响应被完整发送前保持 waiting 状态, 阻止连接被关闭
	 */
	waitingReply wait.Wait

	uploading atomic.AtomicBool

	// 标记客户端是否在发送指令
	sending atomic.AtomicBool

	// 客户端正在发送的数据量
	exceptedArgCount uint32

	// 已经接受的参数数量, 即len(args)
	receivedCount uint32

	// 已经接受到的命令参数, 每个参数有一个 []byte 表示
	args [][]byte
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	c.conn.Close()
	return nil
}
