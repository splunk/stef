package traces

import (
	"bytes"

	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/tigrannajaryan/stef/stef-go/pkg"
	"github.com/tigrannajaryan/stef/stef-otel/oteltef"
	"github.com/tigrannajaryan/stef/stef-pdata/internal/otlptools"
)

type PdataToSTEFTraces struct {
	tempAttrs oteltef.Attributes
	otlp2tef  otlptools.Otlp2Tef
	Sorted    bool
}

func (d *PdataToSTEFTraces) WriteTraces(src ptrace.Traces, writer *oteltef.SpansWriter) error {
	otlp2tef := &d.otlp2tef

	if d.Sorted {
		src.ResourceSpans().Sort(
			func(a, b ptrace.ResourceSpans) bool {
				return otlptools.CmpResourceSpans(a, b) < 0
			},
		)

		for i := 0; i < src.ResourceSpans().Len()-1; {
			if otlptools.CmpResourceSpans(src.ResourceSpans().At(i), src.ResourceSpans().At(i+1)) == 0 {
				src.ResourceSpans().At(i + 1).ScopeSpans().MoveAndAppendTo(src.ResourceSpans().At(i).ScopeSpans())
				j := 0
				src.ResourceSpans().RemoveIf(
					func(metrics ptrace.ResourceSpans) bool {
						j++
						return j == i+2
					},
				)
			} else {
				i++
			}
		}
	}

	for i := 0; i < src.ResourceSpans().Len(); i++ {
		rmm := src.ResourceSpans().At(i)
		otlp2tef.ResourceUnsorted(writer.Record.Resource(), rmm.Resource(), rmm.SchemaUrl())

		if d.Sorted {
			rmm.ScopeSpans().Sort(
				func(a, b ptrace.ScopeSpans) bool {
					return otlptools.CmpScopeSpans(a, b) < 0
				},
			)

			for i := 0; i < rmm.ScopeSpans().Len()-1; {
				if otlptools.CmpScopeSpans(rmm.ScopeSpans().At(i), rmm.ScopeSpans().At(i+1)) == 0 {
					rmm.ScopeSpans().At(i + 1).Spans().MoveAndAppendTo(rmm.ScopeSpans().At(i).Spans())
					j := 0
					rmm.ScopeSpans().RemoveIf(
						func(s ptrace.ScopeSpans) bool {
							j++
							return j == i+2
						},
					)
				} else {
					i++
				}
			}
		}

		for j := 0; j < rmm.ScopeSpans().Len(); j++ {
			smm := rmm.ScopeSpans().At(j)
			otlp2tef.ScopeUnsorted(writer.Record.Scope(), smm.Scope(), smm.SchemaUrl())

			if d.Sorted {
				sortSpans(smm.Spans())
			}

			for k := 0; k < smm.Spans().Len(); k++ {
				m := smm.Spans().At(k)
				span2span(m, writer.Record.Span(), otlp2tef)
				if err := writer.Write(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func sortSpans(spans ptrace.SpanSlice) {
	spans.Sort(
		func(a, b ptrace.Span) bool {
			t1 := a.TraceID()
			t2 := b.TraceID()
			c := bytes.Compare(t1[:], t2[:])
			if c > 0 {
				return true
			}
			if c < 0 {
				return false
			}

			s1 := a.ParentSpanID()
			s2 := b.ParentSpanID()
			c = bytes.Compare(s1[:], s2[:])
			if c > 0 {
				return true
			}
			if c < 0 {
				return false
			}

			return a.StartTimestamp() < b.StartTimestamp()
		},
	)
}

func span2span(src ptrace.Span, dst *oteltef.Span, otlp2tef *otlptools.Otlp2Tef) {
	dst.SetTraceID(pkg.Bytes(src.TraceID().String()))
	dst.SetSpanID(pkg.Bytes(src.SpanID().String()))
	dst.SetParentSpanID(pkg.Bytes(src.ParentSpanID().String()))
	dst.SetName(src.Name())
	dst.SetFlags(uint64(src.Flags()))
	dst.SetStartTimeUnixNano(uint64(src.StartTimestamp()))
	dst.SetEndTimeUnixNano(uint64(src.EndTimestamp()))
	dst.SetKind(uint64(src.Kind()))
	dst.SetTraceState(src.TraceState().AsRaw())
	otlp2tef.MapUnsorted(src.Attributes(), dst.Attributes())

	dst.Status().SetCode(uint64(src.Status().Code()))
	dst.Status().SetMessage(src.Status().Message())

	dst.Events().EnsureLen(src.Events().Len())
	for i := 0; i < src.Events().Len(); i++ {
		event2event(src.Events().At(i), dst.Events().At(i), otlp2tef)
	}

	dst.Links().EnsureLen(src.Links().Len())
	for i := 0; i < src.Links().Len(); i++ {
		link2link(src.Links().At(i), dst.Links().At(i), otlp2tef)
	}
}

func link2link(src ptrace.SpanLink, dst *oteltef.Link, otlp2tef *otlptools.Otlp2Tef) {
	otlp2tef.MapUnsorted(src.Attributes(), dst.Attributes())
	dst.SetFlags(uint64(src.Flags()))
	dst.SetTraceState(src.TraceState().AsRaw())
	dst.SetTraceID(pkg.Bytes(src.TraceID().String()))
	dst.SetSpanID(pkg.Bytes(src.SpanID().String()))
}

func event2event(src ptrace.SpanEvent, dst *oteltef.Event, otlp2tef *otlptools.Otlp2Tef) {
	dst.SetName(src.Name())
	dst.SetTimeUnixNano(uint64(src.Timestamp()))
	otlp2tef.MapUnsorted(src.Attributes(), dst.Attributes())
}
