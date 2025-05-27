package main

import(
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	name string
}

var(
	clients = make(map[net.Conn]Client)
	broadcast = make(chan string)
	mutex sync.Mutex
)

func main(){
	listener,err:= net.Listen("tcp",":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	
	fmt.Println("Server started on :8080")

	go handleBroadcast()
	for{
		conn,err := listener.Accept()
		if err!=nil {
			fmt.Println("Connection error:",err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn){
	defer conn.Close()
	conn.Write([]byte("Enter your name: "))
	nameReader := bufio.NewReader(conn)
	name,_ := nameReader.ReadString('\n')
	name=strings.TrimSpace(name)

	client := Client{conn:conn,name:name}

	mutex.Lock()
	clients[conn]=client
	mutex.Unlock()

	broadcast <- fmt.Sprintf("%s joined the chat", name)
	reader:= bufio.NewReader(conn)
	for{
		msg,err := reader.ReadString('\n')
		if err != nil{
			mutex.Lock()
			delete(clients,conn)
			mutex.Unlock()
			broadcast <- fmt.Sprintf("%s left the chat", name)
				return
		}

		message := fmt.Sprintf("[%s]: [%s]",name,strings.TrimSpace(msg))
		broadcast <- message
	}

}

func handleBroadcast(){
	for{
		msg := <- broadcast
		fmt.Println(msg)

		mutex.Lock()
		for conn:= range clients{
			_,err := fmt.Fprintln(conn,msg)
			if err != nil{
				conn.Close()
				delete(clients,conn)
			}
		}
		mutex.Unlock()
	}
}
