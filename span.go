package gostellar

import (
	"fmt"
	sPb "github.com/c12s/scheme/stellar"
	"github.com/golang/protobuf/proto"
	"github.com/rs/xid"
	"strings"
	"time"
)

type SpanContext struct {
	s *sPb.SpanContext
}

func (s *SpanContext) String() string {
	return fmt.Sprintf("Ctx:(tid: %s sid: %s pid: %s) ", s.s.TraceId, s.s.SpanId, s.s.ParrentSpanId)
}

func NewSpanContext(tid, pid string) *SpanContext {
	return &SpanContext{
		&sPb.SpanContext{
			TraceId:       tid,
			SpanId:        xid.New().String(),
			ParrentSpanId: pid,
			Baggage:       map[string]string{},
		},
	}
}

type Span struct {
	s *sPb.Span
}

func InitSpan(c *SpanContext, name string) *Span {
	return &Span{
		&sPb.Span{
			SpanContext: c.s,
			Name:        name,
			Logs:        map[string]string{},
			Tags:        map[string]string{},
		},
	}
}

func (s *Span) Child(name string) *Span {
	context := NewSpanContext(s.s.SpanContext.TraceId, s.s.SpanContext.ParrentSpanId)
	span := InitSpan(context, name)
	defer span.StartTime()
	return span
}

func (s *Span) AddLog(kvs ...*KV) {
	for _, kv := range kvs {
		s.s.Logs[kv.Key] = kv.Value
	}
}

func (s *Span) AddTag(kvs ...*KV) {
	for _, kv := range kvs {
		s.s.Tags[kv.Key] = kv.Value
	}
}

func (s Span) AddBaggage(kvs ...*KV) {
	for _, kv := range kvs {
		s.s.SpanContext.Baggage[kv.Key] = kv.Value
	}
}

func (s *Span) StartTime() {
	s.s.StartTime = time.Now().UnixNano() / mil //milliseconds
}

func (s *Span) EndTime() {
	s.s.EndTime = time.Now().UnixNano() / mil // milliseconds
}

func (s *Span) Finish() {
	s.EndTime()
	//send to colledtor
	fmt.Println(fmt.Sprintf("Span[%s].Finish() %d ms", s.s.SpanContext.SpanId, (s.s.EndTime - s.s.StartTime)))
	data, err := s.Marshall()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = Log(data, s.s.SpanContext.TraceId, s.s.SpanContext.SpanId)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (span *Span) Serialize() *Values {
	s := map[string][]string{}
	s[trace_id] = []string{span.s.SpanContext.TraceId}
	s[span_id] = []string{span.s.SpanContext.SpanId}
	s[parrent_span_id] = []string{span.s.SpanContext.ParrentSpanId}
	s[tags] = []string{span.digestTags()}
	return &Values{md: s}
}

func (s *Span) ingestTags(existing string) {
	for _, pair := range strings.Split(existing, tag_sep) {
		val := strings.Split(pair, pair_sep)
		s.AddTag(&KV{Key: val[0], Value: val[1]})
	}
}

func (s *Span) digestTags() string {
	t := []string{}
	for k, v := range s.s.Tags {
		t = append(t, fmt.Sprintf("%s:%s", k, v))
	}
	return strings.Join(t, tag_sep)
}

func (s *Span) String() string {
	return fmt.Sprintf("Span: %s %s", s.s.SpanContext, s.s.Name)
}

func (s *Span) Marshall() ([]byte, error) {
	return proto.Marshal(s.s)
}
