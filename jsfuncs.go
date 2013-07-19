// jsfuncs.go
package main

import (
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"net/http"
	"sms77"
)

type jsIRCStruct struct {
	Privmsg, Action, Notice, Join, Part, ChangeNick func(call otto.FunctionCall) otto.Value
	Channel, Server, Nick                           string
	//AdditionalChannels []string
}

var jsIRC *jsIRCStruct

func RegisterJSFuncs() {
	js.Set("123", func(call otto.FunctionCall) otto.Value {
		return otto.UndefinedValue()
	})

	js.Set("testobj", struct {
		A, B func(call otto.FunctionCall) otto.Value
	}{
		A: func(call otto.FunctionCall) otto.Value {
			return otto.UndefinedValue()
		},
		B: func(call otto.FunctionCall) otto.Value {
			return otto.TrueValue()
		},
	})

	jsIRC = &jsIRCStruct{
		Privmsg: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				Irccon.Privmsg(call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Action: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				Irccon.Privmsg(call.Argument(0).String(), "\001ACTION "+call.Argument(1).String()+"\001")
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Notice: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				Irccon.Notice(call.Argument(0).String(), call.Argument(1).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Join: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				Irccon.Join(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Part: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				Irccon.Part(call.Argument(0).String())
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		ChangeNick: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				Irccon.Nick(call.Argument(0).String())
				jsIRC.Nick = call.Argument(0).String()
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		Channel: Channel,
		Server:  Server,
		Nick:    Nickname,
	}

	js.Set("IRC", jsIRC)

	js.Set("DB", struct {
		Authenticate, SaveToDB, ExistsInDB, GetRandomFromDB, GetAndDeleteFirstFromDB, GetNamedFromDB, SaveNamedToDB func(call otto.FunctionCall) otto.Value
	}{
		Authenticate: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				val, _ := js.ToValue(Authenticate(call.ArgumentList[0].String(), call.ArgumentList[1].String()))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		SaveToDB: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsObject() {
				val, err := call.ArgumentList[1].Export()
				if err != nil {
					return otto.FalseValue()
				}

				SaveToDB(call.ArgumentList[0].String(), val.(map[string]interface{}))
				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
		ExistsInDB: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsObject() {
				val, err := call.ArgumentList[1].Export()
				if err != nil {
					return otto.FalseValue()
				}

				val1, _ := js.ToValue(ExistsInDB(call.ArgumentList[0].String(), val.(map[string]interface{})))
				return val1
			} else {
				return otto.FalseValue()
			}
		},
		GetRandomFromDB: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				rand, err := GetRandomFromDB(call.ArgumentList[0].String())

				if err != nil {
					return otto.FalseValue()
				}

				val, _ := js.ToValue(rand)

				return val
			} else {
				return otto.FalseValue()
			}
		},
		GetAndDeleteFirstFromDB: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				rand, err := GetAndDeleteFirstFromDB(call.ArgumentList[0].String())

				if err != nil {
					return otto.FalseValue()
				}

				val, _ := js.ToValue(rand)

				return val
			} else {
				return otto.FalseValue()
			}
		},
		GetNamedFromDB: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				item, err := GetNamedFromDB(call.ArgumentList[0].String(), call.ArgumentList[1].String())

				if err != nil {
					return otto.FalseValue()
				}

				val, _ := js.ToValue(item)

				return val
			} else {
				return otto.FalseValue()
			}
		},
		SaveNamedToDB: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsObject() {
				val, err := call.ArgumentList[1].Export()
				if err != nil {
					return otto.FalseValue()
				}

				err = SaveNamedToDB(call.ArgumentList[0].String(), val.(map[string]interface{}))

				if err != nil {
					return otto.FalseValue()
				}

				return otto.TrueValue()
			} else {
				return otto.FalseValue()
			}
		},
	})

	js.Set("Lib", struct {
		HttpGet, B func(call otto.FunctionCall) otto.Value
	}{
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

				val, _ := js.ToValue(ret)

				return val
			} else {
				return otto.FalseValue()
			}
		},
		B: func(call otto.FunctionCall) otto.Value {
			return otto.TrueValue()
		},
	})

	js.Set("SMS", struct {
		Send, Balance, Status, Flipdebug, GetPhonebook, SetPhonebook, DelPhonebook func(call otto.FunctionCall) otto.Value
	}{
		Send: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
				Sms := &sms77.Sms{}
				Sms.Text = call.Argument(1).String()
				Sms.To = call.Argument(0).String()
				val, _ := js.ToValue(sms77.Sendsms(Sms))
				return val
			} else {
				return otto.FalseValue()
			}
		},
		Balance: func(call otto.FunctionCall) otto.Value {
			val, _ := js.ToValue(sms77.Balance())
			return val
		},
		Status: func(call otto.FunctionCall) otto.Value {
			if len(call.ArgumentList) == 1 && call.ArgumentList[0].IsString() {
				val, _ := js.ToValue(sms77.SmsStatus(call.Argument(0).String()))
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
				val, _ := js.ToValue(sms77.GetPhonebookEntries(srch))
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
	})

}
