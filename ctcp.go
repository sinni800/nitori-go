package main

import (
	//"code.google.com/p/go.net/html"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"path"
	"strconv"
	"strings"
	"time"
)

func (i *instance) DCCSend(nick string, file string) error {

	//if _, ok := i.DCCHandshakes[nick]; !ok {

	//inch := make(chan string, 5)

	ip := net.ParseIP(conf.ExternalIP)
	var ipn uint32
	for _, b := range ip {
		ipn <<= 8
		ipn += uint32(b)
	}

	listen, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0})
	if err != nil {
		return errors.New("Couldnt listen to socket")
	}

	_, port, _ := net.SplitHostPort(listen.Addr().String())

	f, err := conf.FileSystemPrefixes[conf.DCCFilesystem].Dir.Open(file)

	if err != nil {
		return errors.New("File does not exist")
	}

	finfo, _ := f.Stat()

	//i.DCCHandshakes[nick] = inch
	i.Irc.Privmsg(nick, "\x01DCC SEND "+path.Base(strings.Trim(file, " "))+" "+strconv.Itoa(int(ipn))+" "+port+" "+strconv.FormatInt(finfo.Size(), 10))

	go func() {
		/*
			select {
			case msg := <-inch:
				spl := strings.Split(msg, " ")
				if spl[0] == "DCC" && spl[1] == "ACCEPT" {
					if
				}
			case <-time.After(8 * time.Minute):
				delete(i.DCCHandshakes, nick)
				return
			}
		*/
		defer listen.Close()
		deadline := time.Now().Add(8 * time.Minute)
		listen.SetDeadline(deadline)
		con, err := listen.Accept()

		if err != nil {
			return
		}

		go io.Copy(ioutil.Discard, con)
		io.Copy(con, f)
	}()
	//}
	return nil
}
