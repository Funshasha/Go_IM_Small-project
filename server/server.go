package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User //在线用户列表
	mapLock   sync.RWMutex
	Message   chan string //消息广播的channel
}

//创建一个 server 的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//监听 Message 广播消息 channel 的 goroutine , 一旦有消息就发送给全部的在线 user
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		//将 msg 发送给全部在线的 User
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

//广播消息的方法
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	//fmt.Println("链接建立成功...")
	user := NewUser(conn, s)
	user.Online()

	//监听用户是否活跃的 channel
	isLive := make(chan bool)

	//接受客户端传递发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err", err)
				return
			}
			//提取用户消息的信息（去除'\n'）
			msg := string(buf[:n-1])

			//将得到的消息进行广播
			user.DoMessage(msg)

			//用户的任意消息，代表当前用户是一个活跃用户
			isLive <- true
		}
	}()
	//编写一个定时强踢
	for {
		select {
		case <-isLive:
			//当前用户是活跃的， 应该重置定时器
			//为了激活下面的 case 定时器
		case <-time.After(time.Second * 300):
			//已经超时  将当前的客户端的 User 强制关闭
			user.SendMsg("you go out!!!")

			//销毁用户资源

			close(user.C)

			//关闭连接
			conn.Close()

			//退出当前的 Handel
			return //or runtime.Goexit()
		}
	}
}

//启动服务器的一个接口
func (s *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()
	//启动监听 message 的 goroutine
	go s.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err", err)
			continue
		}
		//do handler
		go s.Handler(conn)
	}

}
