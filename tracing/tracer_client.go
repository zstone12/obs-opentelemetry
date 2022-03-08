// Copyright 2022 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var _ stats.Tracer = (*clientTracer)(nil)

type clientTracer struct {
	config            *config
	histogramRecorder map[string]metric.Float64Histogram
}

func newClientOption(opts ...Option) (client.Option, *config) {
	cfg := newConfig(opts)
	ct := &clientTracer{config: cfg}

	ct.createMeasures()

	return client.WithTracer(ct), cfg
}

func (c *clientTracer) createMeasures() {
	clientDurationMeasure, err := c.config.meter.NewFloat64Histogram(ClientDuration)
	handleErr(err)

	c.histogramRecorder = map[string]metric.Float64Histogram{
		ClientDuration: clientDurationMeasure,
	}
}

func (c *clientTracer) Start(ctx context.Context) context.Context {
	var spanName string
	if c.config.spanNameFormatter != nil {
		spanName = c.config.spanNameFormatter(ctx)
	}

	ri := rpcinfo.GetRPCInfo(ctx)
	ctx, _ = c.config.tracer.Start(ctx, spanName,
		oteltrace.WithTimestamp(getStartTimeOrDefault(ri, time.Now())),
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)

	return ctx
}

func (c *clientTracer) Finish(ctx context.Context) {
	span := oteltrace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Stats().Level() == stats.LevelDisabled {
		return
	}

	st := ri.Stats()
	rpcStart := st.GetEvent(stats.RPCStart)
	rpcFinish := st.GetEvent(stats.RPCFinish)
	duration := rpcFinish.Time().Sub(rpcStart.Time())
	elapsedTime := float64(duration) / float64(time.Millisecond)

	attrs := []attribute.KeyValue{
		RPCSystemKitex,
		RequestProtocolKitex,
		semconv.RPCMethodKey.String(ri.To().Method()),
		semconv.RPCServiceKey.String(ri.To().ServiceName()),
		RPCSystemKitexRecvSize.Int64(int64(st.RecvSize())),
		RPCSystemKitexSendSize.Int64(int64(st.SendSize())),
	}

	// The source operation dimension maybe cause high cardinality issues
	if c.config.recordSourceOperation {
		attrs = append(attrs, SourceOperationKey.String(ri.From().Method()))
	}

	span.SetAttributes(attrs...)

	injectStatsEventsToSpan(span, st)

	if st.Error() != nil {
		attrs = append(attrs, StatusKey.String(codes.Error.String()))
		RecordErrorSpan(span, st.Error(), c.config.withStackTrace, attrs...)
	}

	span.End(oteltrace.WithTimestamp(getEndTimeOrDefault(ri, time.Now())))

	metricsAttributes := extractMetricsAttributesFromSpan(span)
	c.histogramRecorder[ClientDuration].Record(ctx, elapsedTime, metricsAttributes...)
}
