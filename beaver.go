package main

import (
	"crypto/tls"
	"fmt"
	irc "github.com/fluffle/goirc/client"
)


func handlePRIVMSG(conn *irc.Conn, line *irc.Line) {
    if line.Public() {
        channel := line.Target()
        text := line.Text()
        // line.Time
        fmt.Printf("%s:%s:%s\n", channel, line.Nick, text)
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
