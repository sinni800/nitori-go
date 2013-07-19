package main

import (
	"flag"
	"fmt"
	"github.com/thoj/go-ircevent"
	"os"
	//"strconv"
	//"container/vector"
	"strings"
)

var Irccon *irc.Connection

var fNickname *string = flag.String("nick", "nitori", "Nickname")
var fNicknameregex *string = flag.String("nickregex", "natori(?:-chan)?", "Nick Regex")
var fServer *string = flag.String("host", "irc.freenode.net:6667", "IRC server")
var fChannel *string = flag.String("channel", "##rlc", "IRC channel")
var fNickservpwd *string = flag.String("nickservpwd", "meruto", "Nickserv password")

var Nickname string
var Nicknameregex string
var Server string
var Channel string
var Nickservpwd string

//var AdditionalChannels []string

func init() {

	flag.Parse()
	Nickname = *fNickname
	Nicknameregex = *fNicknameregex
	Server = *fServer
	Channel = *fChannel
	Nickservpwd = *fNickservpwd
	//AdditionalChannels = make([]string, 0, 0)
}

func IrcGetMessageSource(e *irc.Event) string {
	if len(e.Arguments) == 0 {
		return e.Message
	}
	if strings.HasPrefix(e.Arguments[0], "#") {
		return e.Arguments[0]
	}
	return e.Nick
}

func InitIRC() {

	Irccon = irc.IRC(Nickname, "natori-gobot")

	Irccon.AddCallback("001", func(e *irc.Event) {
		Irccon.Join(Channel)
		Irccon.Privmsg("nickserv", "identify "+Nickservpwd)
	})

	Irccon.AddCallback("JOIN", func(e *irc.Event) {
		//Irccon.Privmsg(Channel, "Nick: "+e.Nick+" Source: "+e.Source+" GetMsgSource: "+e.GetMessageSource()+" Message: "+e.Message)
		raise("join", e.Nick, IrcGetMessageSource(e), "")
	})

	Irccon.AddCallback("PART", func(e *irc.Event) {
		raise("part", e.Nick, e.Arguments[0], "")
	})

	Irccon.AddCallback("QUIT", func(e *irc.Event) {
		raise("quit", e.Nick, "", "")
	})

	/*Irccon.AddCallback("MODE", func(e *irc.Event) {
		raise("mode", e.Nick, e.Arguments[0], e.Arguments[1] + " " + e.Arguments[2])
	})*/

	//nicklist
	Irccon.AddCallback("353", func(e *irc.Event) {

	})

	Irccon.AddCallback("PRIVMSG", func(e *irc.Event) {
		//raise("quit", e.Nick, e.GetMessageSource(), e.Message)

		if e.Message[0] == '!' {
			call := strings.SplitN(e.Message, " ", 2)
			raise(call[0], e.Nick, IrcGetMessageSource(e), e.Message[len(call[0]):])
		}

		raise("privmsg", e.Nick, IrcGetMessageSource(e), e.Message)

		raiseRegex(e.Nick, IrcGetMessageSource(e), e.Message)

	})

	err := Irccon.Connect(Server)
	if err != nil {
		fmt.Printf("%s\n", err)
		fmt.Printf("%#v\n", Irccon)
		os.Exit(1)
	}

}
