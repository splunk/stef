package sortedbyresource

import (
	"io"
	"slices"

	"modernc.org/b/v2"

	"github.com/splunk/stef/go/otel/otelstef"
)

type SortedTree struct {
	byResource *b.Tree[*otelstef.Resource, *ByResource]
}

type ByResource struct {
	byScope *b.Tree[*otelstef.Scope, *ByScope]
}

type ByScope struct {
	logs LogRecords
}

type LogRecords []*otelstef.LogRecord

func NewSortedLogs() *SortedTree {
	return &SortedTree{
		byResource: b.TreeNew[*otelstef.Resource, *ByResource](otelstef.CmpResource),
	}
}

func (s *SortedTree) ByResource(resource *otelstef.Resource) *ByResource {
	byResource, exists := s.byResource.Get(resource)
	if !exists {
		byResource = &ByResource{
			byScope: b.TreeNew[*otelstef.Scope, *ByScope](otelstef.CmpScope),
		}
		s.byResource.Set(resource, byResource)
	}
	return byResource
}

func (r *ByResource) ByScope(scope *otelstef.Scope) *ByScope {
	byScope, exists := r.byScope.Get(scope)
	if !exists {
		byScope = &ByScope{}
		r.byScope.Set(scope, byScope)
	}
	return byScope
}

func (s *ByScope) Append(log *otelstef.LogRecord) {
	s.logs = append(s.logs, log)
}

func (s *SortedTree) SortValues() {
	_ = s.Iter(func(_ *otelstef.Resource, resource *ByResource) error {
		return resource.Iter(func(_ *otelstef.Scope, scope *ByScope) error {
			slices.SortFunc(scope.logs, otelstef.CmpLogRecord)
			return nil
		})
	})
}

func (s *SortedTree) ToStef(writer *otelstef.LogsWriter) error {
	return s.Iter(func(resource *otelstef.Resource, byResource *ByResource) error {
		writer.Record.SetResource(resource)
		return byResource.Iter(func(scope *otelstef.Scope, byScope *ByScope) error {
			writer.Record.SetScope(scope)
			for _, log := range byScope.logs {
				writer.Record.Log().CopyFrom(log)
				if err := writer.Write(); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func (s *SortedTree) Iter(f func(*otelstef.Resource, *ByResource) error) error {
	iter, err := s.byResource.SeekFirst()
	if err != nil {
		return nil
	}
	for {
		key, value, err := iter.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := f(key, value); err != nil {
			return err
		}
	}
}

func (r *ByResource) Iter(f func(*otelstef.Scope, *ByScope) error) error {
	iter, err := r.byScope.SeekFirst()
	if err != nil {
		return nil
	}
	for {
		key, value, err := iter.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := f(key, value); err != nil {
			return err
		}
	}
}
