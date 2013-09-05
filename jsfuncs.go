// jsfuncs.go
package main

import (
	"github.com/robertkrimen/otto"
	"io/ioutil"
	//"labix.org/v2/mgo/bson"
	"net/http"
	"sms77"
)

type jsIRCStruct struct {
	Privmsg, Action, Notice, Join, Part, ChangeNick, DCCSend func(call otto.FunctionCall) otto.Value
	Channel, Server, Nick                                    string
	//AdditionalChannels []string
}

type jsDBStruct struct {
	Authenticate, SaveToDB, ExistsInDB, GetFromDB, GetRandomFromDB,
	GetAndDeleteFirstFromDB, GetNamedFromDB, SaveNamedToDB,
	DeleteFromDB, Reconnect func(call otto.FunctionCall) otto.Value
	LastError string
}

type jsLibStruct struct {
	HttpGet, IsAuthenticated func(call otto.FunctionCall) otto.Value
	Conf                     *config
}

type jsSMSStruct struct {
	Send, Balance, Status, Flipdebug, GetPhonebook, SetPhonebook, DelPhonebook func(call otto.FunctionCall) otto.Value
}

func (i *instance) RegisterJSFuncs() {
	i.Jsobj = &struct {
		IRC *jsIRCStruct
		DB  *jsDBStruct
		Lib *jsLibStruct
		SMS *jsSMSStruct
	}{}

	i.js.Set("123", func(call otto.FunctionCall) otto.Value {
		return otto.UndefinedValue()
	})

	i.js.Set("testobj", struct {
		A, B func(call otto.FunctionCall) otto.Value
	}{
		A: func(call otto.FunctionCall) otto.Value {
			return otto.UndefinedValue()
		},
		B: func(call otto.FunctionCall) otto.Value {
			return otto.TrueValue()
		},
	})

	i.Jsobj.IRC = &jsIRCStruct{
		Privmsg: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				i.Irc.Privmsg(call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Action: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				i.Irc.Privmsg(call.Argument(0).String(), "\001ACTION "+call.Argument(1).String()+"\001")
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Notice: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				i.Irc.Notice(call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Join: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				i.Irc.Join(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Part: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				i.Irc.Part(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		ChangeNick: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				i.Irc.Nick(call.Argument(0).String())
				i.Jsobj.IRC.Nick = call.Argument(0).String()
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		DCCSend: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				if err := i.DCCSend(call.ArgumentList[0].String(), call.ArgumentList[1].String()); err == nil {
					return otto.TrueValue()
				} else {
					val, _ := i.js.ToValue(err)
					return val
				}
			} else {
				return otto.FalseValue()
			}
		},
		Channel: i.Irccfg.Channel,
		Server:  i.Irccfg.Host,
		Nick:    i.Irccfg.Nick,
	}

	i.js.Set("IRC", i.Jsobj.IRC)

	i.Jsobj.DB = &jsDBStruct{
		Authenticate: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				val, _ := i.js.ToValue(Authenticate(call.ArgumentList[0].String(), call.ArgumentList[1].String()))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		SaveToDB: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsObject() {
				val, err := call.ArgumentList[1].Export()
				if err != nil {
					return otto.UndefinedValue()
				}

				err = SaveToDB(call.ArgumentList[0].String(), val.(map[string]interface{}))
				if err != nil {
					str, _ := i.js.ToValue(err.Error())
					return str
				}
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		ExistsInDB: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsObject() {
				val, err := call.ArgumentList[1].Export()

				if err != nil {
					return otto.UndefinedValue()
				}

				val1, _ := i.js.ToValue(ExistsInDB(call.ArgumentList[0].String(), val.(map[string]interface{})))
				return val1
			} else {
				return otto.FalseValue()
			}
		},
		GetFromDB: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsObject() {
				val, err := call.ArgumentList[1].Export()

				if err != nil {
					return otto.UndefinedValue()
				}

				val1, err := GetFromDB(call.ArgumentList[0].String(), val.(map[string]interface{}))

				if err != nil {
					i.Jsobj.DB.LastError = err.Error()
					return otto.UndefinedValue()
				}

				valret := make([]map[string]interface{}, 0, 0)

				for _, bsonval := range val1 {
					valret = append(valret, map[string]interface{}(bsonval))
				}

				val2, err := i.js.ToValue([]map[string]interface{}(valret))

				if err != nil {
					return otto.UndefinedValue()
				}

				return val2
			} else {
				return otto.FalseValue()
			}
		},
		GetRandomFromDB: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				rand, err := GetRandomFromDB(call.ArgumentList[0].String())

				if err != nil {
					i.Jsobj.DB.LastError = err.Error()
					return otto.UndefinedValue()
				}

				val, _ := i.js.ToValue(map[string]interface{}(rand))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		GetAndDeleteFirstFromDB: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				rand, err := GetAndDeleteFirstFromDB(call.ArgumentList[0].String())

				if err != nil {
					i.Jsobj.DB.LastError = err.Error()
					return otto.UndefinedValue()
				}

				val, _ := i.js.ToValue(map[string]interface{}(rand))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		GetNamedFromDB: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				item, err := GetNamedFromDB(call.ArgumentList[0].String(), call.ArgumentList[1].String())

				if err != nil {
					i.Jsobj.DB.LastError = err.Error()
					return otto.UndefinedValue()
				}

				val, _ := i.js.ToValue(map[string]interface{}(item))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		SaveNamedToDB: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsObject() {
				val, err := call.ArgumentList[1].Export()
				if err != nil {
					i.Jsobj.DB.LastError = err.Error()
					return otto.UndefinedValue()
				}

				err = SaveNamedToDB(call.ArgumentList[0].String(), val.(map[string]interface{}))

				if err != nil {
					i.Jsobj.DB.LastError = err.Error()
					return otto.UndefinedValue()
				}

				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		DeleteFromDB: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsObject() {
				val, err := call.ArgumentList[1].Export()
				if err != nil {
					i.Jsobj.DB.LastError = err.Error()
					return otto.UndefinedValue()
				}

				err = DeleteFromDB(call.ArgumentList[0].String(), val.(map[string]interface{}))

				if err != nil {
					i.Jsobj.DB.LastError = err.Error()
					return otto.UndefinedValue()
				}

				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Reconnect: func(call otto.FunctionCall) otto.Value {
			i.Jsobj.DB.LastError = ""
			if err := reconnectDB(); err == nil {
				return otto.TrueValue()
			} else {
				i.Jsobj.DB.LastError = err.Error()
				return otto.UndefinedValue()
			}
		},
		LastError: "",
	}

	i.js.Set("DB", i.Jsobj.DB)

	i.Jsobj.Lib = &jsLibStruct{
		HttpGet: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				resp, err := http.Get(call.ArgumentList[0].String())

				if err != nil {
					return otto.FalseValue()
				}

				body, err := ioutil.ReadAll(resp.Body)

				resp.Body.Close()

				if err != nil {
					return otto.FalseValue()
				}

				ret := struct {
					Body, Statusstring string
					Status             int
					Header             http.Header
				}{
					Body:         string(body),
					Status:       resp.StatusCode,
					Statusstring: resp.Status,
					Header:       resp.Header,
				}

				val, _ := i.js.ToValue(ret)

				return val
			} else {
				return otto.FalseValue()
			}
		},
		IsAuthenticated: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				if _, ok := i.Authenticatednicks[call.Argument(0).String()]; ok {
					return otto.TrueValue()
				}
			}
			return otto.FalseValue()
		},
		Conf: conf,
	}

	i.js.Set("Lib", i.Jsobj.Lib)

	i.Jsobj.SMS = &jsSMSStruct{
		Send: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				Sms := &sms77.Sms{}
				Sms.Text = call.Argument(1).String()
				Sms.To = call.Argument(0).String()
				val, _ := i.js.ToValue(sms77.Sendsms(Sms))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		Balance: func(call otto.FunctionCall) otto.Value {
			val, _ := i.js.ToValue(sms77.Balance())
			return val
		},
		Status: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				val, _ := i.js.ToValue(sms77.SmsStatus(call.Argument(0).String()))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		Flipdebug: func(call otto.FunctionCall) otto.Value {
			if sms77.Debug {
				sms77.Debug = false
				return otto.FalseValue()
			} else {
				sms77.Debug = true
				return otto.TrueValue()
			}
		},
		GetPhonebook: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 4 && call.ArgumentList[0].IsString() &&
				call.ArgumentList[1].IsString() &&
				call.ArgumentList[2].IsString() &&
				call.ArgumentList[3].IsString() {
				srch := sms77.PhonebookEntry{}
				srch.Id = call.Argument(0).String()
				srch.Nick = call.Argument(1).String()
				srch.Empfaenger = call.Argument(2).String()
				srch.Email = call.Argument(3).String()
				val, _ := i.js.ToValue(sms77.GetPhonebookEntries(srch))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		SetPhonebook: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 4 && call.ArgumentList[0].IsString() &&
				call.ArgumentList[1].IsString() &&
				call.ArgumentList[2].IsString() &&
				call.ArgumentList[3].IsString() {
				srch := sms77.PhonebookEntry{}
				srch.Id = call.Argument(0).String()
				srch.Nick = call.Argument(1).String()
				srch.Empfaenger = call.Argument(2).String()
				srch.Email = call.Argument(3).String()
				if sms77.EditOrNewPhonebookEntry(srch) == nil {
					return otto.TrueValue()
				} else {
					return otto.FalseValue()
				}
			} else {
				return otto.FalseValue()
			}
		},
		DelPhonebook: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				if sms77.DelPhonebookEntry(call.Argument(0).String()) == nil {
					return otto.TrueValue()
				} else {
					return otto.FalseValue()
				}
			} else {
				return otto.FalseValue()
			}
		},
	}

	i.js.Set("SMS", i.Jsobj.SMS)

}
