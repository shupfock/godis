package tcp

import (
	"bufio"
	"context"
	"godis/src/lib/logger"
	"godis/src/lib/sync/atomic"
	"godis/src/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

type Client struct {
	// tcp 连接
	Conn net.Conn
	// 当服务端开始发送数据时进入 waiting 状态, 阻止其他 goroutine 关闭连接
	Waiting wait.Wait
}

func (c *Client) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	c.Conn.Close()
	return nil
}

type EchoHandler struct {
	// 保存所有工作状态client 的集合, 把map 当 set 用
	// 需要使用并发安全容器
	activeConn sync.Map
	// 和 tcp service 中相同的关闭状态标志位
	closing atomic.AtomicBool
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (eh *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if eh.closing.Get() {
		conn.Close()
	}

	client := &Client{
		Conn: conn,
	}
	eh.activeConn.Store(client, 1)

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				eh.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		// 发送前置为 waiting 状态
		client.Waiting.Add(1)

		// 模拟关闭时未完成发送的情况
		// logger.Info("sleep")
		// time.Sleep(10 * time.Second)
		b := []byte(msg)
		conn.Write(b)

		// 发送完毕, 结束 waiting
		client.Waiting.Done()
	}
}

func (eh *EchoHandler) Close() error {
	logger.Info("handler shutind down...")
	eh.closing.Set(true)

	eh.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*Client)
		client.Close()
		return true
	})
	return nil
}
