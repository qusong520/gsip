package sip

import (
	"bytes"
	"fmt"
	"strings"
)

type MessageID string

// Message introduces common SIP message RFC 3261 - 7.
type Message interface {
	MessageID() MessageID
	Short() string
	String() string
	Method() RequestMethod

	// SIP 请求是根据起始行中的 Request-Line 来区分的
	// Request-Line = Method SP Request-URI SP SIP-VERSION CRLF
	StartLine() string

	// SIP 版本 SIP/2.0 必须发送大写
	SipVersion() string
	SetSipVersion(version string)

	// 头部参数集合
	Headers() []Header
	GetHeaderString(name string) []string
	// 获取头部参数
	GetHeaders(name string) []Header
	// 设置头部参数(自动格式化)
	AddHeaderString(headName string, value string) error
	// 设置头部参数
	AddHeader(header Header)
	// 替换头部参数
	ReplaceHeader(header Header)
	PrependHeader(header Header)
	// 将参数插入至某个头部后面
	PrependHeaderAfter(header Header, afterName string)
	// 删除头部参数
	DelHeader(name ...string)

	// Body returns message body.
	Body() string
	// SetBody sets message body.
	SetBody(body string, setContentLength bool)

	// CallID returns 'Call-ID' header.
	// CallID() (*CallID, bool)
	CallID() *CallID
	// Via returns the top 'Via' header field.
	Via() (ViaHeader, bool)
	// ViaHop returns the first segment of the top 'Via' header.
	ViaHop() (*ViaHop, bool)
	// From returns 'From' header field.
	From() *FromHeader
	// To returns 'To' header field.
	To() *ToHeader
	// CSeq returns 'CSeq' header field.
	CSeq() *CSeq
	// Expires returns 'Expires' header field.
	Expires() *Expires
	// Authorization returns 'Authorization' header field.
	Authorization() *Authorization

	ContentLength() *ContentLength
	ContentType() *ContentType
	Contact() *ContactHeader

	Transaction() Transaction      // 返回事务层指针
	SetTransaction(tx Transaction) // 返回事务层指针
	Transport() string             // 传输层
	Source() string                // 来源地址
	SetSource(src string)          // 设置源地址
	Destination() string           // 目的地地址
	SetDestination(dest string)    // 设置目的地地址

	IsCancel() bool   // 是否关闭
	IsAck() bool      // 是否是 ACK 信息
	DialogId() string // 对话(Dialog)
}

// basic message implementation
type message struct {
	// message headers
	*headers
	tx         Transaction
	messID     MessageID
	sipVersion string
	body       string
	startLine  func() string
	src        string
	dest       string
}

func (msg *message) MessageID() MessageID {
	return msg.messID
}

func (msg *message) StartLine() string {
	return msg.startLine()
}

func (msg *message) String() string {
	var buffer bytes.Buffer

	// write message start line
	buffer.WriteString(msg.StartLine() + "\r\n")
	// Write the headers.
	buffer.WriteString(msg.headers.String())
	// message body
	buffer.WriteString("\r\n" + msg.Body())

	return buffer.String()
}

func (msg *message) SipVersion() string {
	return msg.sipVersion
}

func (msg *message) SetSipVersion(version string) {
	msg.sipVersion = version
}

func (msg *message) Body() string {
	return msg.body
}

// SetBody sets message body, calculates it length and add 'Content-Length' header.
func (msg *message) SetBody(body string, setContentLength bool) {
	msg.body = body
	if setContentLength {
		hers := msg.GetHeaders("Content-Length")
		if len(hers) == 0 {
			length := ContentLength(len(body))
			msg.AddHeader(&length)
		} else {
			length := ContentLength(len(body))
			hers[0] = &length
		}
	}
}

func (msg *message) Transport() string {
	if viaHop, ok := msg.ViaHop(); ok {
		return viaHop.Transport
	} else {
		return DefaultProtocol
	}
}

func (msg *message) Transaction() Transaction {
	return msg.tx
}

func (msg *message) SetTransaction(tx Transaction) {
	msg.tx = tx
}

func (msg *message) Source() string {
	return msg.src
}
func (msg *message) SetSource(src string) {
	msg.src = src
}
func (msg *message) Destination() string {
	return msg.dest
}
func (msg *message) SetDestination(dest string) {
	msg.dest = dest
}
func (msg *message) DialogId() string {
	var (
		callId  string
		fromTag MaybeString
		toTag   MaybeString
	)
	if callHead := msg.CallID(); callHead != nil {
		callId = string(*callHead)
	}
	if from := msg.From(); from != nil {
		fromTag, _ = from.Params.Get("tag")
	}
	if to := msg.To(); to != nil {
		toTag, _ = to.Params.Get("tag")
	}
	return fmt.Sprintf("%s#%s#%s", callId, fromTag, toTag)
}

func CopyHeaders(name string, from, to Message) {
	name = strings.ToLower(name)
	for _, h := range from.GetHeaders(name) {
		to.AddHeader(h.Copy())
	}
}
