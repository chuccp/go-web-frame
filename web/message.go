package web

type Message struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

func (msg *Message) IsOK() bool {
	return msg.Code == 200
}

func Ok(msg ...string) *Message {
	if len(msg) > 0 {
		return &Message{
			Code: 200,
			Msg:  msg[0],
		}
	}
	return &Message{
		Code: 200,
		Msg:  "ok",
	}
}
func Data(data any) *Message {
	return &Message{
		Code: 200,
		Msg:  "ok",
		Data: data,
	}
}
func DataType(t string, data any) *Message {
	return &Message{
		Type: t,
		Code: 200,
		Msg:  "ok",
		Data: data,
	}
}
func DataCode(code int, data any) *Message {
	return &Message{
		Code: code,
		Msg:  "ok",
		Data: data,
	}
}

func ErrorMessage(msg ...string) *Message {
	m := "error"
	if len(msg) > 0 {
		m = msg[0]
	}
	return &Message{
		Code: 500,
		Msg:  m,
	}
}
func Error(err ...error) *Message {
	m := "error"
	if len(err) > 0 {
		m = err[0].Error()
	}
	return &Message{
		Code: 500,
		Msg:  m,
	}
}
func Errors(data any, msg ...error) *Message {
	m := "error"
	if len(msg) > 0 {
		m = msg[0].Error()
	}
	return &Message{
		Code: 500,
		Msg:  m,
		Data: data,
	}
}
func Unauthorized(data any, msg ...error) *Message {
	m := "error"
	if len(msg) > 0 {
		m = msg[0].Error()
	}
	return &Message{
		Code: 401,
		Msg:  m,
		Data: data,
	}
}
