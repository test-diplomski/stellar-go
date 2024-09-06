package gostellar

import (
	"context"
	"errors"
	sPb "github.com/c12s/scheme/stellar"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
)

func FromRequest(r *http.Request, name string) (Spanner, error) {
	traceId := r.Context().Value(trace).(*Values).Get(trace_id)[0]
	spanId := r.Context().Value(trace).(*Values).Get(span_id)[0]
	tags := r.Context().Value(trace).(*Values).Get(tags)[0] //k:v;kv;...;kv:kv
	if traceId != "" && spanId != "" {
		span := InitSpan(NewSpanContext(traceId, spanId), name)
		defer span.StartTime()

		if tags != "" {
			span.ingestTags(tags)
		}
		return span, nil
	}
	return nil, errors.New("No trace context in request")
}

func FromContext(ctx context.Context, name string) (Spanner, error) {
	traceId := ctx.Value(trace).(*Values).Get(trace_id)[0]
	spanId := ctx.Value(trace).(*Values).Get(span_id)[0]
	tags := ctx.Value(trace).(*Values).Get(tags)[0] //k:v;kv;...;kv:kv
	if traceId != "" && spanId != "" {
		span := InitSpan(NewSpanContext(traceId, spanId), name)
		defer span.StartTime()

		if tags != "" {
			span.ingestTags(tags)
		}
		return span, nil
	}
	return nil, errors.New("No trace in context")
}

func FromGRPCContext(ctx context.Context, name string) (Spanner, error) {
	// Read metadata from client.
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		traceId := md[trace_id][0]
		spanId := md[span_id][0]
		tags := md[tags][0] //k:v;kv;...;kv:kv
		if traceId != "" && spanId != "" {
			span := InitSpan(NewSpanContext(traceId, spanId), name)
			defer span.StartTime()

			if tags != "" {
				span.ingestTags(tags)
			}
			return span, nil
		}
	}
	return nil, errors.New("No trace in context")
}

func FromCustomSource(ctx *sPb.SpanContext, tags map[string]string, name string) (Spanner, error) {
	// Read metadata from client.
	if ctx != nil {
		span := InitSpan(&SpanContext{ctx}, name)
		defer span.StartTime()

		for k, v := range tags {
			span.AddTag(&KV{k, v})
		}

		return span, nil
	}
	return nil, errors.New("No trace in custom source")
}

func FromResponse(w http.ResponseWriter, name string) (Spanner, error) {
	traceId := w.Header().Get(trace_id)
	spanId := w.Header().Get(span_id)
	// parrentSpanId := w.Header().Get(parrent_span_id)
	tags := w.Header().Get(tags)
	if traceId != "" && spanId != "" {
		span := InitSpan(NewSpanContext(traceId, spanId), name)
		defer span.StartTime()

		if tags != "" {
			span.ingestTags(tags)
		}
		return span, nil
	}
	return nil, errors.New("No trace in context")
}

func NewTracedRequest(method, url string, body io.Reader, span Spanner) (*http.Request, error) {
	c := context.WithValue(context.Background(), trace, span.Serialize())
	return http.NewRequestWithContext(c, method, url, body)
}

func NewTracedContext(ctx context.Context, span Spanner) context.Context {
	if ctx != nil {
		return context.WithValue(ctx, trace, span.Serialize())
	} else {
		return context.WithValue(context.Background(), trace, span.Serialize())
	}
}

func NewTracedGRPCContext(ctx context.Context, span Spanner) context.Context {
	if ctx != nil {
		return metadata.NewOutgoingContext(ctx, span.Serialize().md)
	} else {
		return metadata.NewOutgoingContext(context.Background(), span.Serialize().md)
	}
}

func TracedRequest(r *http.Request, span Spanner) *http.Request {
	c := context.WithValue(context.Background(), trace, span.Serialize())
	return r.WithContext(c)
}

func TracedResponse(w http.ResponseWriter, span Spanner) http.ResponseWriter {
	s := span.Serialize()
	w.Header().Set(trace_id, s.Get(trace_id)[0])
	w.Header().Set(span_id, s.Get(span_id)[0])
	w.Header().Set(parrent_span_id, s.Get(parrent_span_id)[0])
	w.Header().Set(tags, s.Get(tags)[0])
	return w
}
