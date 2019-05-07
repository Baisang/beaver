package main

import (
	"crypto/tls"
    "encoding/json"
	"fmt"
	irc "github.com/fluffle/goirc/client"
)

type Message struct {
    Channel string
    Nick string
    Text string
    Time int64
}

func handlePRIVMSG(conn *irc.Conn, line *irc.Line) {
    if line.Public() {
        message := Message{
            Channel: line.Target(),
            Nick: line.Nick,
            Text: line.Text(),
            Time: line.Time.Unix(),
        }
        blob, _ := json.Marshal(message)
        fmt.Printf(string(blob))
    }
}

func main() {
	cfg := irc.NewConfig("beaver-bot")
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: "irc.ocf.berkeley.edu"}
	cfg.Server = "irc.ocf.berkeley.edu:6697"
	cfg.NewNick = func(n string) string { return n + "^" }
    c := irc.Client(cfg)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		conn.Join("#test")
	})

	c.HandleFunc(irc.PRIVMSG, handlePRIVMSG)

	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		quit <- true 
	})
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}
	<-quit
}
