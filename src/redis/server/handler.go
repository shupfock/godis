package server

import (
	"fmt"
	"context"
	"bufio"
	"io"
	"strconv"
	"net"
	"sync"
	"godis/src/lib/logger"
	"godis/src/redis/reply"
	"godis/src/lib/sync/atomic"
)

var (
	UnknowErrReplyBytes = []byte("-ERR unknow\r\n")
)

type Handler struct {
	/*
	 * 记录活跃的客户端链接
	 * 类型为 *Client -> placeholder
	 */
	activeConn sync.Map

	// db db.DB

	closing  atomic.AtomicBool
}

func NewHandler() *Handler{
	return &Handler{}
}

func (h *Handler)closeClient(client * Client){
	client.Close()
	h.activeConn.Delete(client)
}

func (h *Handler)Handle(ctx context.Context, conn net.Conn){
	if h.closing.Get() {
		// 关闭过程中不接受新链接
		conn.Close()
	}

	// 初始化客户端状态
	client := &Client{
		conn: conn,
	}

	h.activeConn.Store(client, 1)
	reader := bufio.NewReader(conn)

	var fixedLen int64 = 0 // 将要读取的BulkString 的长度
	var err error
	var msg []byte
	for {
		// 读取下一行
		if fixedLen == 0 {  // 正常模式使用 CRLF 区分数据行
			msg, err = reader.ReadBytes('\n')
			// if err != nul {
			// 	if err == io.EOF ||
			// 	err == if.ErrUnexpectedEOF ||
			// 	strings.Contains(er.Error(), "user of closed network connection") {
			// 		logger.Info("connection close")
			// 	} else {
			// 		logger.Warn(client)
			// 	}
			// 	h.closeCilent(client)
			// 	return 
			// }

			// 判断是否以 \r\n 结尾
			if len(msg) == 0 || msg[len(msg)-2] != '\r' {
				errReply := &reply.ProtocolErrReply{Msg: "invalid multibulk legth"}
				client.conn.Write(errReply.ToBytes())
			}
		} else {
			// 当读到第二行, 根据给出的长度进行读取
			msg = make([]byte, fixedLen+2)
			_, err = io.ReadFull(reader, msg)
			// if err != nil {
			// 	if err == io.EOF ||
			// 	err == io.ErrUnexpectedEOF ||
			// 	string.Contains(err.Error, "use of closed netword connection") {
			// 		logger.Info("connect close")
			// 	} else {
			// 		logger.Warn(err)
			// 	}

			// 	h.closeClient(client)
			// 	return 
			// }

			// 判断是否以 \r\n 结尾
			if len(msg) == 0 ||
			msg[len(msg)-2] != '\r' ||
			msg[len(msg)-1] != '\n' {
				errReply := &reply.ProtocolErrReply{Msg: "invalid multibulk length"}
				client.conn.Write(errReply.ToBytes())
			}

			//  读取完毕, 重新使用正常模式
			fixedLen = 0
		}

		if err != nil {
			if err == io.EOF ||
			err == io.ErrUnexpectedEOF {
				logger.Info("connect close")
			} else {
				logger.Warn(err)
			}

			h.closeClient(client)
			return 
		}

		// 解析收到的数据
		if !client.uploading.Get(){
			
			// sending == false 表明收到一条指令
			if msg[0] == '*' {
				exceptedLine, err := strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
				if err != nil {
					client.conn.Write(UnknowErrReplyBytes)
					continue
				}
				// 初始化客户端状态
				client.waitingReply.Add(1)  // 有指令未完成, 阻止服务器关闭
				client.uploading.Set(true)  // 正在接受指令中

				// 初始化计数器和缓冲区
				client.exceptedArgCount = uint32(exceptedLine)
				client.receivedCount = 0
				client.args = make([][]byte, exceptedLine)
			} else {

			}
		} else {
			// 收到了指令的剩余部分
			line := msg[0:len(msg)-2]  // 溢出换行符
			if line[0] == '$' {
				// BulkString 的首行, 读取 string 的长度
				fixedLen, err := strconv.ParseInt(string(line[1:]), 10, 64)
				if err != nil {
					errReply := &reply.ProtocolErrReply{Msg: err.Error()}
					client.conn.Write(errReply.ToBytes())
				}
				if fixedLen <= 0 {
					errReply := &reply.ProtocolErrReply{Msg: "invalid multibulk length"}
					client.conn.Write(errReply.ToBytes())
				}
			} else {
				// 首单参数 
				client.args[client.receivedCount] = line
				client.receivedCount++
			}

			// 一条命令发送完毕
			if client.receivedCount == client.exceptedArgCount {
				client.uploading.Set(false)

				// 执行命令并响应
				fmt.Println(client.args)

				// 重置客户端状态
				client.exceptedArgCount = 0
				client.receivedCount = 0
				client.args = nil
				client.waitingReply.Done()

			}
		}


	}
}

func (h *Handler) Close() error{
	logger.Info("handler shuting down ...")
	h.closing.Set(true)

	h.activeConn.Range(func(key interface{}, val interface{}) bool{
		client := key.(*Client)
		client.Close()
		return true
	})
	return nil
}