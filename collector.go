package gostellar

import (
	"context"
	"fmt"
	sPb "github.com/c12s/scheme/stellar"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"time"
)

type NatsCollector struct {
	natsConnection *nats.Conn
	topic          string
}

func NewCollector(address, topic string) (*NatsCollector, error) {
	natsConnection, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}

	return &NatsCollector{
		natsConnection: natsConnection,
		topic:          topic,
	}, nil
}

func (n NatsCollector) Collect(data []*sPb.Span) {
	b, err := proto.Marshal(&sPb.LogBatch{data})
	if err != nil {
		return
	}
	n.natsConnection.Publish(n.topic, b)
}

type Deamon struct {
	Path string
	m    Collector
}

func InitCollector(path string, c Collector) (*Deamon, error) {
	err := CheckCollectorDir(path)
	if err != nil {
		return nil, err
	}

	return &Deamon{
		Path: path,
		m:    c,
	}, nil
}

func (dc *Deamon) Start(ctx context.Context, t time.Duration) {
	ticker := time.NewTicker(t)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Collector done")
			return
		case <-ticker.C:
			go func() {
				batch := []*sPb.Span{}
				for elem := range CollectTraces(ctx, dc.Path) {
					switch value := elem.(type) {
					case error:
						fmt.Println(value.Error())
					case *sPb.Span:
						batch = append(batch, value)
					default:
						fmt.Println(value)
						return
					}
				}
				dc.m.Collect(batch) // send to service
				ClearDir(dc.Path)   // clear current dir
			}()
		}
	}
}
