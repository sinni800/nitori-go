// main
package main

import (
	"flag"
	"fmt"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"launchpad.net/goyaml"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sms77"
	"strings"
	"time"
)

var end chan bool = make(chan bool, 1)
var conf *config
var Configpath *string = flag.String("config", "config.yaml", "Path to the yaml config file")

type instance struct {
	Name  string     `yaml:"-"`
	js    *otto.Otto `yaml:"-"`
	Jsobj *struct {
		IRC *jsIRCStruct
		DB  *jsDBStruct
		Lib *jsLibStruct
		SMS *jsSMSStruct
	} `yaml:"-"`
	Pluginfuncs map[string]map[string]otto.Value                                     `yaml:"-"`
	Plugindocs  map[string]map[string]string                                         `yaml:"-"`
	Corefuncs   map[string]func(src messagesource, line string, parameters []string) `yaml:"-"`
	Irc         *irc.Connection                                                      `yaml:"-"`
	Irccfg      struct {
		UseIrc      bool     `yaml:"useIRC"`
		Host        string   `yaml:"host"`
		Channel     string   `yaml:"primarychannel"`
		Channels    []string `yaml:"channels"`
		Nick        string   `yaml:"nick"`
		Nickregex   string   `yaml:"nickregex"`
		Nickservpwd string   `yaml:"nickservpwd"`
	} `yaml:"irc"`
	Irclog             map[string]*ChannelLogStack
	Authenticatednicks map[string]bool `yaml:"-"`
	DCCHandshakes      DCCHands
}

type DCCHands map[string]chan string

type DCCState int

const (
	DCCWaitingForAccept = iota
	DCCWaitingForConnection
	DCCTransfering
)

type DCCFile struct {
	Listener net.Listener
	State    DCCState
	Conn     net.TCPConn
	ch       chan string
	//Timeout
}

type Filesystemhttp struct {
	Dir                   http.Dir `yaml:"path"`
	ReadNeedsAuth         bool     `yaml:"readAuth"`
	ListNeedsAuth         bool     `yaml:"listAuth"`
	ZipNeedsAuth          bool     `yaml:"zipAuth"`
	UploadNeedsAuth       bool     `yaml:"uploadAuth"`
	CreateFolderNeedsAuth bool     `yaml:"createFolderAuth"`
	DeleteNeedsAuth       bool     `yaml:"deleteAuth"`
}

type config struct {
	Instances          map[string]*instance      `yaml:"configs"`
	FileSystemPrefixes map[string]Filesystemhttp `yaml:"filesystems"`

	DocInstance   string `yaml:"docInstance"`
	SMSDebug      bool   `yaml:"smsdebug"`
	Plugindir     string `yaml:"plugindir`
	DCCFilesystem string `yaml:"dccFilesystem"`
	ExternalIP    string `yaml:"externalIP"`

	Mongo struct {
		UseMongo      bool   `yaml:"useMongo"`
		MongoAddress  string `yaml:"mongoAddress"`
		MongoDatabase string `yaml:"mongoDatabase"`
	} `yaml:"mongo"`

	Web struct {
		Http          bool   `yaml:"http"`
		Https         bool   `yaml:"https"`
		Httplisten    string `yaml:"httplisten"`
		Httpslisten   string `yaml:"httpslisten"`
		HttpsCertFile string `yaml:"httpscertfile"`
		HttpsKeyFile  string `yaml:"httpskeyfile"`
		Hostname      string `yaml:"hostname"`
		Templatedir   string `yaml:"templatedir"`
	}
}

func (i *instance) Authenticated(nick string) bool {
	_, authed := i.Authenticatednicks[nick]
	return authed
}

func init() {
	flag.Parse()
	conf = &config{}
}

func main() {

	b, err := ioutil.ReadFile(*Configpath)

	if err == nil {
		err = goyaml.Unmarshal(b, &conf)
	}

	if err != nil {
		println("couldnt read config file: " + err.Error())
		os.Exit(1)
	}

	sms77.Debug = conf.SMSDebug

	/* DB */

	fmt.Println(conf)

	if conf.Mongo.UseMongo {
		connectDB()

		go func() {
			for {
				time.Sleep(5 * time.Minute)
				reconnectDB()
			}
		}()
	}

	/* HTTP */

	if conf.Web.Http || conf.Web.Https {
		handleHttpFuncs()

		if conf.Web.Http {
			fmt.Println("Http active on " + conf.Web.Httplisten)
			go func() {
				err := http.ListenAndServe(conf.Web.Httplisten, nil)
				if err != nil {
					panic(err.Error())
				}
			}()
		}
		if conf.Web.Https {
			go func() {
				err := http.ListenAndServeTLS(conf.Web.Httpslisten, conf.Web.HttpsCertFile, conf.Web.HttpsKeyFile, nil)
				if err != nil {
					panic(err.Error())
				}
			}()
		}
	}

	/* HTTP FileSystem */

	/* Instances */

	for key, value := range conf.Instances {
		value.Name = key
		value.Pluginfuncs = make(map[string]map[string]otto.Value)
		value.Plugindocs = make(map[string]map[string]string)
		value.Corefuncs = make(map[string]func(source messagesource, line string, parameters []string))
		value.Authenticatednicks = make(map[string]bool)
		value.js = otto.New()
		value.RegisterJSFuncs()
		value.RegisterIRCCoreFuncs()
		value.Irclog = make(map[string]*ChannelLogStack)
		value.DCCHandshakes = make(DCCHands)

		filepath.Walk(conf.Plugindir, func(path string, info os.FileInfo, err error) error {
			if strings.ToLower(filepath.Ext(path)) == ".js" {
				value.loadplugin(path)
			}
			return nil
		})

		if value.Irccfg.UseIrc {
			value.InitIRC()

			go func(inst *instance) {
				for {
					inst.Irc.Loop()
					time.Sleep(20 * time.Second)
					inst.Irc.Reconnect()
				}
			}(value)
		}
	}

	<-end
}

type messagesource struct {
	Nick    string
	Source  string
	Channel string
}
