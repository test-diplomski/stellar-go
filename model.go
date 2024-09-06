package gostellar

import (
	"context"
	sPb "github.com/c12s/scheme/stellar"
	"time"
)

type Spanner interface {
	Child(name string) *Span
	AddLog(kvs ...*KV)
	AddTag(kvs ...*KV)
	AddBaggage(kvs ...*KV)
	StartTime()
	EndTime()
	Finish() // send data to collector and maybe serialize to ctx ot request
	Serialize() *Values
	Marshall() ([]byte, error)
}

type Tracer interface {
	Span(name string) Spanner
	Finish()
}

type Scanner interface {
	Start(ctx context.Context, t time.Duration)
}

type Collector interface {
	Collect(data []*sPb.Span)
}

type KV struct {
	Key   string
	Value string
}

type Values struct {
	md map[string][]string
}

func (v Values) Get(key string) []string {
	return v.md[key]
}
