package sip

type Transaction interface {
	Origin() Request
	// Session() Session
	String() string
	Errors() <-chan error
	Done() <-chan bool
}

type ServerTransaction interface {
	Transaction
	SendResponse(res Response) error
	Requests() <-chan Request
}

type ClientTransaction interface {
	Transaction
	Responses() <-chan Response
}
