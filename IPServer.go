// TCPServer project main.go
package main

/**
本程序用于给位置隐私保护程序建立转发树使用
*/
import (
	"bytes"
	"encoding/binary"
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

// 消息类型
const IP_SERVER_MSG = 0

// 运行模式
const NC_MODE_STRING = "NC模式"
const OD_MODE_STRING = "OD模式"

var runMode = NC_MODE_STRING

// 处理到来消息
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

	printTreeMsg()

	// 运行模式   NC模式 和 OD模式

	// 返回数据的格式   运行模式  + 节点分配来的序号 + 父节点IP
	ret := runMode + "," + strconv.Itoa(index) + "," + fatherIp
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

		result := solveSocketMsg(rAddr)

		_, err1 := conn.Write(int32tobytes(IP_SERVER_MSG))
		data := []byte(result)
		_, err1 = conn.Write(int32tobytes(int32(len(data))))
		_, err1 = conn.Write(data)
		if err1 != nil {
			continue
		}

		// 处理与client的交互
		//go handleClient(conn)

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
		// 先发送消息类型
		//_, err2 := conn.Write(int32tobytes(0))
		//if err2 != nil {
		//	return
		//}
		// 发送数据长度
		data := []byte(result)

		_, err1 := conn.Write(int32tobytes(IP_SERVER_MSG))
		_, err1 = conn.Write(int32tobytes(int32(len(data))))

		_, err1 = conn.Write(data)
		if err1 != nil {
			return
		}
	}
}

func printTreeMsg() {
	for i, temp := range ipTable {
		if temp != "" {
			fmt.Println(strconv.Itoa(i) + "	" + temp)
		}
	}
}

// 将int32转化为大端序的bytes数组
func int32tobytes(arg int32) []byte {
	s1 := make([]byte, 0)
	buf := bytes.NewBuffer(s1)

	// 数字转 []byte, 网络字节序为大端字节序
	binary.Write(buf, binary.BigEndian, arg)

	return buf.Bytes()
}

// 异常处理
func checkErr1(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
