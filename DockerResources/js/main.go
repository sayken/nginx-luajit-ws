package main

import (
	"strconv"
	"fmt"
	"net"
	"os"
)

// CheckError checks for errors
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func main() {
	/*
		nginxからのproxyを受ける->nginx上のproxyで受けてしまうと、同じポートで送出できない。
		ので、nginxでのudp proxyは不要ということになるが、うーん。まあようはlistenしなければいいんだよね。

		あるいはtokenを返すのに使うか。

		proxyしてることで、このソケットからのoutが特殊なことになってくれるのでは？という期待がある。
	*/
	


	// receiver
	// ここでudpのポートが決定されてしまうのでは？感があるが、ここから返したらどうなるんだろう->proxyしてる部分に対して綺麗に返る。なるほど。
	// streamの先にはいろいろ書けるので、おおーー分散できるね。自分に到達できるキーを返せばいい、と。で、外部に見せるポートは8080唯一つになる。
	ServerAddr, err := net.ResolveUDPAddr("udp", ":8081")
	

	CheckError(err)

	// start listen.
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	
	CheckError(err)
	

	defer ServerConn.Close()


	// 受け取りバッファ
	buf := make([]byte, 1024)
	
	// 適当な無限ループ
	for {
		// bufのクリアが必要

		n, addr, err := ServerConn.ReadFromUDP(buf)
		// fmt.Println("addr:", addr)

		if err != nil {
			fmt.Println("error: ", err)
			continue
		}

		// 1byte目が特定のマークだったら、データを分解して特定の宛先へと転送する。
		// addrでローカルからだったら、とかやってもいいと思う。
		if buf[0] == 'd' {
			count, err := strconv.Atoi(string(buf[1:3]))
			if err != nil {
				fmt.Println("error1: ", err)
				continue
			}

			ipAndPort := string(buf[3:count + 3])// get port and id
			
			host, portStr, err := net.SplitHostPort(ipAndPort);
			if err != nil {
				fmt.Println("error2: ", err)
				continue
			}
			
			port, err := strconv.Atoi(portStr)
			if err != nil {
				fmt.Println("error3: ", err)
				continue
			}

			targetAddr := net.UDPAddr{IP:net.ParseIP(host), Port:port, Zone:"sample"}

			data := buf[3+count:n]
			// fmt.Println("data:", string(data), "vs len:", len(data), "and index", 3+count, "total message len:", n)

			// send udp data to target ip:port.
			ServerConn.WriteTo(data, &targetAddr)
			continue
		}

		addrStr := addr.IP.String() + ":" + strconv.Itoa(addr.Port)

		// ポート番号を返す。
		ServerConn.WriteTo([]byte(addrStr), addr)
	}
}
