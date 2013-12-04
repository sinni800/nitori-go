package main

import (
	//"fmt"
	"github.com/thoj/go-ircevent"
	//"os"
	//"strconv"
	//"container/vector"
	"errors"
	"strings"
	"time"
)

type ChannelLogStack struct {
	channel     string
	stack       []IrcChannelLogMessage
	stackLength int
	stackStart  int
}

func (c *ChannelLogStack) Full() bool {
	return c.stackLength == len(c.stack)
}

func (c *ChannelLogStack) Push(msg IrcChannelLogMessage) error {
	if c.stackLength == len(c.stack) {
		return errors.New("Max length reached")
	}
	c.stackLength++
	c.stack[(c.stackStart+c.stackLength-1)%len(c.stack)] = msg
	return nil
}

func (c *ChannelLogStack) Pop() (ret IrcChannelLogMessage) {
	if c.stackLength == 0 {
		ret = IrcChannelLogMessage{}
		return
	}
	ret = c.stack[c.stackStart]
	c.stackStart++
	if c.stackStart > (len(c.stack) - 1) {
		c.stackStart = 0
	}

	c.stackLength--
	return
}

func (c *ChannelLogStack) PushPopFull(msg IrcChannelLogMessage) (retmsg IrcChannelLogMessage) {
	if c.Full() {
		retmsg = c.Pop()
	}
	c.Push(msg)
	return
}

func NewChannelLogStack(channel string) *ChannelLogStack {
	c := &ChannelLogStack{}
	c.channel = channel
	c.stack = make([]IrcChannelLogMessage, 200)
	c.stackLength = 0
	c.stackStart = 0
	return c
}

func (c *ChannelLogStack) Out() (stuff []IrcChannelLogMessage) {
	stuff = make([]IrcChannelLogMessage, 0, len(c.stack))

	for x := 0; x < c.stackLength; x++ {
		//fmt.Println("putting " + strconv.Itoa((LastTenChannelMessagesStart + x) % 10) + " into " + strconv.Itoa(x))
		stuff = append(stuff, c.stack[(c.stackStart+x)%len(c.stack)])
	}

	return
}

type IrcChannelLogMessage struct {
	Sender  string
	Message string
	T       time.Time
}

func (i *instance) IrcGetMessageSource(e *irc.Event) string {
	if len(e.Arguments) == 0 {
		return e.Message
	}
	if strings.HasPrefix(e.Arguments[0], "#") {
		if e.Arguments[0] != i.Irccfg.Channel {
			return e.Arguments[0]
		}
		return e.Arguments[0]
	}
	return e.Nick
}

func (i *instance) IrcGetMessageChannel(e *irc.Event) string {
	if len(e.Arguments) == 0 {
		return i.Irccfg.Channel
	}

	if strings.HasPrefix(e.Arguments[0], "#") {
		return e.Arguments[0]
	} else {
		return i.Irccfg.Channel
	}
}

func (i *instance) NewMessageSource(e *irc.Event) messagesource {
	return messagesource{e.Nick, i.IrcGetMessageSource(e), i.IrcGetMessageChannel(e)}
}

func (e *instance) InitIRC() {
	i := e

	i.Irclog[i.Irccfg.Channel] = NewChannelLogStack(i.Irccfg.Channel)

	for _, c := range i.Irccfg.Channels {
		i.Irclog[c] = NewChannelLogStack(c)
	}

	i.Irc = irc.IRC(i.Irccfg.Nick, "natori-gobot")

	i.Irc.AddCallback("001", func(e *irc.Event) {
		i.Irc.Privmsg("nickserv", "identify "+i.Irccfg.Nickservpwd)
		i.Irc.Join(i.Irccfg.Channel)
		for _, val := range i.Irccfg.Channels {
			i.Irc.Join(val)
		}

	})

	i.Irc.AddCallback("JOIN", func(e *irc.Event) {
		//Irccon.Privmsg(Channel, "Nick: "+e.Nick+" Source: "+e.Source+" GetMsgSource: "+e.GetMessageSource()+" Message: "+e.Message)
		i.raise("join", messagesource{e.Nick, i.IrcGetMessageSource(e), i.IrcGetMessageChannel(e)}, "", false)
	})

	i.Irc.AddCallback("PART", func(e *irc.Event) {
		if e.Arguments[0] == i.Irccfg.Channel {
			delete(i.Authenticatednicks, e.Nick)
		}
		i.raise("part", messagesource{e.Nick, e.Arguments[0], e.Arguments[0]}, "", false)
	})

	i.Irc.AddCallback("QUIT", func(e *irc.Event) {
		delete(i.Authenticatednicks, e.Nick)
		i.raise("quit", messagesource{e.Nick, "", ""}, "", i.Authenticated(e.Nick))
	})

	/*Irccon.AddCallback("MODE", func(e *irc.Event) {
		raise("mode", e.Nick, e.Arguments[0], e.Arguments[1] + " " + e.Arguments[2])
	})*/

	//nicklist
	i.Irc.AddCallback("353", func(e *irc.Event) {
		//fmt.Println(e)
	})

	i.Irc.AddCallback("NICK", func(e *irc.Event) {
		delete(i.Authenticatednicks, e.Nick)
		i.Authenticatednicks[e.Message] = true
	})

	//whois
	i.Irc.AddCallback("311", func(e *irc.Event) {
		//fmt.Println(e)
	})

	i.Irc.AddCallback("CTCP", func(e *irc.Event) {
		ctcpsplit := strings.Split(e.Message, " ")
		i.raise("ctcp:"+ctcpsplit[0], messagesource{e.Nick, "", ""}, strings.TrimPrefix(e.Message, "CTCP "), i.Authenticated(e.Nick))
	})

	i.Irc.AddCallback("PRIVMSG", func(e *irc.Event) {
		authed := i.Authenticated(e.Nick)
		//raise("quit", e.Nick, e.GetMessageSource(), e.Message)

		if strings.HasPrefix(i.IrcGetMessageSource(e), "#") {
			if val, ok := i.Irclog[i.IrcGetMessageSource(e)]; ok {
				val.PushPopFull(IrcChannelLogMessage{Sender: e.Nick, Message: e.Message, T: time.Now()})
			}
		}

		//Verb
		if len(e.Message) > len(i.Irccfg.Nick) + 1 {
			if strings.HasPrefix(e.Message, i.Irccfg.Nick) {
				if strings.ContainsAny(string(e.Message[len(i.Irccfg.Nick)]), ":,~") {
					withoutnick := e.Message[len(i.Irccfg.Nick)+2:]
					singlewordsplit := strings.SplitN(withoutnick, " ", 2)
					twowordsplit := strings.SplitN(withoutnick, " ", 3)

					if len(singlewordsplit) == 2 {
						i.raise("verb:"+singlewordsplit[0], i.NewMessageSource(e), singlewordsplit[1], authed)
					} else if len(singlewordsplit) == 1 {
						i.raise("verb:"+singlewordsplit[0], i.NewMessageSource(e), "", authed)
					}

					if len(twowordsplit) == 3 {
						i.raise("verb:"+twowordsplit[0]+" "+twowordsplit[1], i.NewMessageSource(e), twowordsplit[2], authed)
					} else if len(twowordsplit) == 2 {
						i.raise("verb:"+twowordsplit[0]+" "+twowordsplit[1], i.NewMessageSource(e), "", authed)
					}
				}
			}
		}

		//Bang-Commands
		if e.Message[0] == '!' {
			call := strings.SplitN(e.Message, " ", 2)
			if len(call) == 2 {
				i.raise(call[0], i.NewMessageSource(e), call[1], authed)
			} else {
				i.raise(call[0], i.NewMessageSource(e), "", authed)
			}

		}

		//Irccon.SendRaw("WHOIS " + e.Nick)
		i.raise("privmsg", i.NewMessageSource(e), e.Message, authed)
		i.raiseRegex(i.NewMessageSource(e), e.Message, authed)
	})

	err := i.Irc.Connect(i.Irccfg.Host)
	if err != nil {
		//return nil
	}

	//return nil
}
