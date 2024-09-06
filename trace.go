package gostellar

import (
	"fmt"
	"github.com/rs/xid"
)

type Trace struct {
	id   string
	name string
}

func (t *Trace) Span(name string) *Span {
	context := NewSpanContext(t.id, "-")
	span := InitSpan(context, name)
	defer span.StartTime()
	return span
}

func (t *Trace) Finish() {
	//send to colledtor
	fmt.Println("Trace.Finish()")
}

func (t *Trace) String() string {
	return fmt.Sprintf("Trace: %s %s", t.id, t.name)
}

func Init(name string) *Trace {
	return &Trace{
		id:   xid.New().String(),
		name: name,
	}
}
