package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户的端口 API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//启动监听当前 User channel 消息的 goroutine
	go user.ListenMessage()
	return user
}

//用户上线业务
func (u *User) Online() {
	//用户上线， 将用户加入到 onlineMap 中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	//广播当前用户上线消息
	u.server.BroadCast(u, "login")

}

//用户下线业务
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	//广播当先用户下线消息
	u.server.BroadCast(u, "offline")
}

func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

//用户处理消息业务
func (u *User) DoMessage(msg string) {

	if msg == "who" {
		//查询当前用户
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMeg := "[" + user.Addr + "]" + user.Name + ":" + "online...\n"
			u.SendMsg(onlineMeg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式： rename| 张三
		newName := strings.Split(msg, "|")[1]

		//判断 name 是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("used username now\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg("you update username:" + u.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式： to|username|消息内容
		// 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("Incorrect message format \"to|username|message.\" \n")
			return
		}

		//根据用户名 得到对方User对象
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("Username does not exist\n")
			return
		}

		//获取消息内容 通过对方的User对象将消息发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("content nil, restart\n")
			return
		}
		remoteUser.SendMsg(u.Name + "to:" + content)

	} else {
		u.server.BroadCast(u, msg)
	}
}

// 监听当前 User channel 的方法，一旦有消息 就直接发送给对应客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))

	}
}
