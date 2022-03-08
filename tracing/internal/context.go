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

package internal

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type traceCarrierContextKeyType struct {
}

var traceCarrierContextKey traceCarrierContextKeyType

type TraceCarrier struct {
	meter             metric.Meter
	tracer            oteltrace.Tracer
	span              oteltrace.Span
	spanNameFormatter func(context.Context) string
}

func WithTraceCarrier(ctx context.Context, tc *TraceCarrier) context.Context {
	return context.WithValue(ctx, traceCarrierContextKey, tc)
}

func TraceCarrierFromContext(ctx context.Context) *TraceCarrier {
	if tc := ctx.Value(traceCarrierContextKey); tc != nil {
		return tc.(*TraceCarrier)
	}

	return nil
}

func (t *TraceCarrier) Meter() metric.Meter {
	return t.meter
}

func (t *TraceCarrier) SetMeter(meter metric.Meter) {
	t.meter = meter
}

func (t *TraceCarrier) Tracer() oteltrace.Tracer {
	return t.tracer
}

func (t *TraceCarrier) SetTracer(tracer oteltrace.Tracer) {
	t.tracer = tracer
}

func (t *TraceCarrier) Span() oteltrace.Span {
	return t.span
}

func (t *TraceCarrier) SetSpan(span oteltrace.Span) {
	t.span = span
}

func (t *TraceCarrier) SpanNameFormatter() func(context.Context) string {
	return t.spanNameFormatter
}

func (t *TraceCarrier) SetSpanNameFormatter(spanNameFormatter func(context.Context) string) {
	t.spanNameFormatter = spanNameFormatter
}
