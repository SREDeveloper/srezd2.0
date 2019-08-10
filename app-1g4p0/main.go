package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"strconv"
	"time"
	"github.com/srezd/app-1g4p0/blockchain"
	"github.com/srezd/app-1g4p0/tcpkeepalive"
)

func checkError(err error,info string) (res bool) {
	
	if(err != nil){
		fmt.Println(info+"  " + err.Error())
		return false
	}
	return true
}
//輸入：0-上链init，1-查询query，2-修改invoke，3-更新channelCC
//writeto:0-init成功，invoke成功，updatecc成功;1-init失败，query失败,updatecc失败，3-invoke失败
func handleConnection(fSetup blockchain.FabricSetup,conn net.Conn) {
    fmt.Printf("Serving %s\n", conn.RemoteAddr().String())
    for {
	data := make([]byte, 4800)
	n, err := conn.Read(data)
	if(checkError(err,"Connection")==false){
		fmt.Printf("Read err00: \n")
		conn.Close()
		break
	}
	sRev := string(data[0:n])
	if n == 0 || err != nil {
		continue
	}
	//messnager := make(chan byte)  
	//心跳计时  
	//go HeartBeating(conn,messnager,200)
	//检测每次Client是否有数据传来
	//go GravelChannel((data[0:n]),messnager)
	//Log("receive data length:",n)
	//Log(conn.RemoteAddr().String(), "receive data string:", sRev)
	fmt.Println("n:", n)
	fmt.Println("sRev:", sRev)
	cmd := strings.SplitN(sRev, "γ", 2)
	fmt.Println("cmd[0]:", cmd[0])
	fmt.Println("cmd[1]:", cmd[1])
	if strings.TrimSpace(cmd[0]) == "0" {
		//Install and instantiate the chaincode
		//response, err = fSetup.Inithoer(strings.TrimSpace(cmd[1]))
		response, err := fSetup.InitSre(cmd[1])
		if err != nil {
			fmt.Printf("Unable to install and instantiate the chaincode: %v\n", err)
			rep := "1"					
			_,err := conn.Write([]byte(rep))
			if(err != nil){					
				fmt.Printf("Write err01: \n")
				conn.Close()
				break
			}
			continue
		} else {
			fmt.Printf("Response from instantiate chaincode %s\n", response)
			rep := "0"
			_,err := conn.Write([]byte(rep))
			if(err != nil){					
				fmt.Printf("Write err02: \n")
				conn.Close()
				break
			}
			continue
		}
	} else if strings.TrimSpace(cmd[0]) == "1" {
		fmt.Printf("strings.TrimSpace(cmd[1]):%s\n", strings.TrimSpace(cmd[1]))
		//time.Sleep(1 * time.Second)
		//cmdstr := cmd[1][0: 32]
		//len0 := len(cmdstr)
		//fmt.Println("len0:", len0)
		response, err := fSetup.QuerySre(strings.TrimSpace(cmd[1]))
		//time.Sleep(1 * time.Second)
		if err != nil {
			fmt.Printf("Unable to query sre on the chaincode: %v\n", err)
			rep := "1"
			_, err := conn.Write([]byte(rep))
			if err != nil {
				fmt.Printf("Write err03: \n")
				conn.Close()
				break
			}
			continue
		} else {
			//response_bytes := []byte(response)
			//response_bytes_len := bytes.Count(response_bytes,nil)					
			response = response + "ʃ"
			response_bytes_len := len(response)
			fmt.Println("response:", response)
			fmt.Println("response_bytes_len:", response_bytes_len)
			tmp := "00000" + strconv.Itoa(response_bytes_len)
			tmp_len := len([]rune(tmp))
			content := tmp[tmp_len-6: tmp_len]
			fmt.Printf("content: %s\n", content)
			_, err := conn.Write([]byte(content))
			if err != nil {
				fmt.Printf("Write err04: \n")
				conn.Close()
				break
			}
			//time.Sleep(1 * time.Second)
			if response == "[]ʃ" {
				fmt.Println("if")
			}else{
				fmt.Println("else")
				_, err = conn.Write([]byte(response))
				if err != nil {
					fmt.Printf("Write err05: \n")
					conn.Close()
					break
				}
			}
			fmt.Printf("Response chaincode: %s\n", response)
			continue
		}
	} else if strings.TrimSpace(cmd[0]) == "2" {
		// Invoke the chaincode
		txId, err := fSetup.InvokeSre(strings.TrimSpace(cmd[1]))
		if err != nil {
			fmt.Printf("Unable to invoke sre on the chaincode: %v\n", err)
			rep := "3"					
			_,err := conn.Write([]byte(rep))
			if(err != nil){
				conn.Close()
				break
			}
			continue
		} else {
			fmt.Printf("Successfully invoke sre, transaction ID: %s\n", txId)
			rep := "0"
			_,err := conn.Write([]byte(rep))
			if(err != nil){
				conn.Close()
				break
			}
			continue
		}
	} else if cmd[0] == "Q" {
		rep := "QQ"
		_,err := conn.Write([]byte(rep))
		if(err != nil){
			conn.Close()
			break
		}
		break
	} else {
		rep := "else"
		_,err := conn.Write([]byte(rep))
		if(err != nil){
			conn.Close()
			break
			}
			continue
		}
	}
	conn.Close()
}

func main() {
	// Definition of the Fabric SDK properties
	fSetup := blockchain.FabricSetup{
		// Network parameters 
		OrdererID: "orderer.hf.srezd.io",
		OrgmspID:  "org4.hf.srezd.io",
		OrdererAdmin:  "Admin",
		OrdererOrgName:   "OrdererOrg",

		// Channel parameters
		ChannelID:     "srezd",
//		ChannelConfig: os.Getenv("GOPATH") + "/src/github.com/srezd/app-1g4p0/fixtures/artifacts/srezd.channel.tx",

		// Chaincode parameters
		ChainCodeID:     "app-1g4",
		ChaincodeGoPath: os.Getenv("GOPATH"),
		ChaincodePath:   "github.com/srezd/app-1g4p0/chaincode/",
		Org_now:         "org4",
		ConfigFile:      "config.yaml",

		// User parameters		
		Org_nowAdmin:        "Admin",
		Org_nowUser: "User1",
	}

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
		return
	}
	// Close SDK
	defer fSetup.CloseSDK()	
	// Install and instantiate the chaincode
	err = fSetup.InstallAndInstantiateCC()
	if err != nil {
		fmt.Printf("Unable to install and instantiate the chaincode: %v\n", err)
		return
	}
	
	listen_sock, err := net.Listen("tcp", ":11111")
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer listen_sock.Close()
	for {
		conn, err := listen_sock.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			break
		}
		fmt.Println("成功连接22！")
		kaConn, _ := tcpkeepalive.EnableKeepAlive(conn)
		kaConn.SetKeepAliveIdle(30*time.Second)
		kaConn.SetKeepAliveCount(4)
		kaConn.SetKeepAliveInterval(5*time.Second)
		defer conn.Close()
		go handleConnection(fSetup,conn)
	}
}