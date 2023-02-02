package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	Conn       net.Conn
	Key        int //当前用户的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		Key:        -1,
	}
	//连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err", err)
		return nil
	}
	client.Conn = conn
	//返回对象
	return client
}

//处理 server 回应的消息 直接显示到标准输出即可
func (c *Client) DealResponse() {

	// for {
	// 	buf := make()
	// 	c.Conn.Read(buf)
	// 	fmt.Println(buf)
	// }
	//一旦client.Conn 有数据，就直接cope 到 stdout 标准输出上，永久阻塞监听
	io.Copy(os.Stdout, c.Conn)
}

func (c *Client) menu() bool {
	var key int
	fmt.Println("1.群发模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.修改用户名")
	fmt.Println("0.退出")

	fmt.Scan(&key)

	if key >= 0 && key <= 3 {
		c.Key = key
		return true
	} else {
		fmt.Println("<<<请输入合法范围的数字>>>")
		return false
	}
}

//查询当前在线用户
func (c *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := c.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("Conn Write err", err)
		return
	}
}

//私聊模式
func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	c.SelectUsers()
	fmt.Println(">>> 请输入聊天对象用户名， exit 退出 <<<")
	fmt.Scan(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>> 请输入消息内容， exit 退出 <<<")
		fmt.Scan(&chatMsg)

		for chatMsg != "exit" {
			//判断消息是否为空
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := c.Conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>> 请输入消息内容， exit 退出 <<<")
			fmt.Scan(&chatMsg)

		}

		c.SelectUsers()
		fmt.Println(">>> 请输入聊天对象用户名， exit 退出 <<<")
		fmt.Scan(&remoteName)

	}
}

func (c *Client) PublicChat() {
	//提示用户输出消息
	var chatMsg string
	fmt.Println(">>> 请输入聊天内容， exit 退出 <<<")
	fmt.Scan(&chatMsg)

	for chatMsg != "exit" {
		//发送消息给服务器端

		//消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.Conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>> 请输入聊天内容， exit 退出 <<<")
		fmt.Scan(&chatMsg)

	}

}

func (c *Client) UpdateName() bool {

	fmt.Println(">>>请输入用户名<<<")
	fmt.Scan(&c.Name)

	sandMsg := "rename|" + c.Name + "\n"
	_, err := c.Conn.Write([]byte(sandMsg))
	if err != nil {
		fmt.Println("client Conn.Write err", err)
		return false
	}
	return true
}

func (c *Client) Run() {
	for c.Key != 0 {
		for c.menu() != true {

		}

		//根据不同的模式处理不同的业务
		switch c.Key {
		case 1:
			//群发模式
			c.PublicChat()
			break
		case 2:
			//私聊模式
			c.PrivateChat()
			break
		case 3:
			//修改用户名
			c.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

// -ip 127.0.0.1
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器ip地址(默认是8888)")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>> 链接服务器失败...")
		return
	}
	//单独开启一个 goroutine 去处理 server 的回传消息
	go client.DealResponse()

	fmt.Println(">>>>>>> 链接服务器成功")

	//启动服务器业务
	client.Run()
}
