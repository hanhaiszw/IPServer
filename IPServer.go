// TCPServer project main.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	// 开启服务器
	openTcpServer(8261)
}

// 处理socket信息

//  0代表是服务器节点
//  下标为子节点id，值为对应的父节点id
//  0     					1     2
//  0   					0     1
//  192.168.137.1   服务器对应的ip
var clientsTree = [3]int{0, 0, 1} // 存储转发树信息
var ipTable = [3]string{"192.168.137.1"}
var clientIndex = 1 // 用户id号
var mutex sync.Mutex

func solveSocketMsg(addr net.Addr) string {
	mutex.Lock()
	defer mutex.Unlock()

	// 取出ip地址
	splits := strings.Split(addr.String(), ":")
	sonIp := splits[0]

	index := -1
	for i, temp := range ipTable {
		if temp == sonIp {
			index = i
		}
	}
	if index == -1 {
		index = clientIndex
		clientIndex++
	}

	ipTable[index] = sonIp

	// 取出父节点ip地址
	fatherIp := ipTable[clientsTree[index]]
	for i, temp := range ipTable {
		if temp != "" {
			fmt.Println(strconv.Itoa(i)+"	"+temp)
		}
	}

	ret := strconv.Itoa(index) + "  " + fatherIp
	return ret
}

func openTcpServer(port int) {
	// 监听8261端口   ":8261"
	strAddr := ":" + strconv.Itoa(port)
	service := strAddr
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkErr1(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkErr1(err)

	fmt.Println("监听端口8261成功，等待client连接")
	// 等待client连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		rAddr := conn.RemoteAddr()

		// rAddr格式为 127.0.0.1:4001
		splits := strings.Split(rAddr.String(), ":")
		//fmt.Println(splits[0] + "已连接")
		log.Print(splits[0] + "已连接")
		// fmt.Println(splits[1])  端口

		// 处理与client的交互
		go handleClient(conn)
	}
}

// 与client的交互
func handleClient(conn net.Conn) {
	defer conn.Close()
	var buf [512]byte
	rAddr := conn.RemoteAddr()
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			//fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
			//fmt.Fprintln(os.Stderr, rAddr.String(), "连接断开")
			//fmt.Println(rAddr.String(), "连接断开")
			log.Println(rAddr.String(), "连接断开")
			return
		}
		fmt.Println("Receive from client", rAddr.String(), string(buf[0:n]))
		result := solveSocketMsg(rAddr)
		_, err2 := conn.Write([]byte(result))
		if err2 != nil {
			return
		}
	}
}

// 异常处理
func checkErr1(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
