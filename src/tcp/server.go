package tcp

/**
 * A tcp server
 */

import (
	"context"
	"fmt"
	"godis/src/lib/logger"
	"godis/src/lib/sync/atomic"
	RedisServer "godis/src/redis/server"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

func ListenAndServe(cfg *Config, handler *RedisServer.Handler) {
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		logger.Fatal(fmt.Sprintf("listen err: %v", err))
	}

	var closing atomic.AtomicBool
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			// 收到信号关闭连接
			logger.Info("shout down...")
			closing.Set(true)
			// 先关闭监听, 阻止连接进入
			listener.Close()
			// 再逐个关闭连接
			handler.Close()
		}
	}()

	logger.Info(fmt.Sprintf("bind: %s, start listening ", cfg.Address))
	defer func() {
		listener.Close()
		handler.Close()
	}()
	ctx, _ := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			if closing.Get() {
				// 收到关闭信号后进入此流程, listener 已被goroutine 提前关闭
				logger.Info("waitting disconnect...")
				wg.Wait()
				return
			}
			logger.Error(fmt.Sprintf("accept err: %v", err))
			continue
		}
		// 创建一个新的协程
		logger.Info("accept link")
		go func() {
			defer func() {
				wg.Done()
			}()
			wg.Add(1)
			handler.Handle(ctx, conn)
		}()
	}
}
