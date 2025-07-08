package services

import (
	protos "github.com/panyam/sdl/gen/go/sdl/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Generator conversions

func toProtoGenerator(g *Generator) *protos.Generator {
	if g == nil {
		return nil
	}

	return &protos.Generator{
		CreatedAt: timestamppb.New(g.CreatedAt),
		UpdatedAt: timestamppb.New(g.UpdatedAt),
		Id:        g.ID,
		CanvasId:  g.CanvasID,
		Name:      g.Name,
		Component: g.Component,
		Method:    g.Method,
		Rate:      g.Rate,
		Duration:  g.Duration,
		Enabled:   g.Enabled,
	}
}

func fromProtoGenerator(p *protos.Generator) *Generator {
	if p == nil {
		return nil
	}

	g := &Generator{
		ID:        p.Id,
		CanvasID:  p.CanvasId,
		Name:      p.Name,
		Component: p.Component,
		Method:    p.Method,
		Rate:      p.Rate,
		Duration:  p.Duration,
		Enabled:   p.Enabled,
	}

	if p.CreatedAt != nil {
		g.CreatedAt = p.CreatedAt.AsTime()
	}
	if p.UpdatedAt != nil {
		g.UpdatedAt = p.UpdatedAt.AsTime()
	}

	return g
}

// Metric conversions

func toProtoMetric(m *Metric) *protos.Metric {
	if m == nil {
		return nil
	}

	return &protos.Metric{
		CreatedAt:         timestamppb.New(m.CreatedAt),
		UpdatedAt:         timestamppb.New(m.UpdatedAt),
		Id:                m.ID,
		CanvasId:          m.CanvasID,
		Name:              m.Name,
		Component:         m.Component,
		Methods:           m.Methods,
		Enabled:           m.Enabled,
		MetricType:        m.MetricType,
		Aggregation:       m.Aggregation,
		AggregationWindow: m.AggregationWindow,
		MatchResult:       m.MatchResult,
		MatchResultType:   m.MatchResultType,
		OldestTimestamp:   m.OldestTimestamp,
		NewestTimestamp:   m.NewestTimestamp,
		NumDataPoints:     m.NumDataPoints,
	}
}

func fromProtoMetric(p *protos.Metric) *Metric {
	if p == nil {
		return nil
	}

	m := &Metric{
		ID:                p.Id,
		CanvasID:          p.CanvasId,
		Name:              p.Name,
		Component:         p.Component,
		Methods:           p.Methods,
		Enabled:           p.Enabled,
		MetricType:        p.MetricType,
		Aggregation:       p.Aggregation,
		AggregationWindow: p.AggregationWindow,
		MatchResult:       p.MatchResult,
		MatchResultType:   p.MatchResultType,
		OldestTimestamp:   p.OldestTimestamp,
		NewestTimestamp:   p.NewestTimestamp,
		NumDataPoints:     p.NumDataPoints,
	}

	if p.CreatedAt != nil {
		m.CreatedAt = p.CreatedAt.AsTime()
	}
	if p.UpdatedAt != nil {
		m.UpdatedAt = p.UpdatedAt.AsTime()
	}

	return m
}

// SystemDiagram conversions

func toProtoSystemDiagram(d *SystemDiagram) *protos.SystemDiagram {
	if d == nil {
		return nil
	}

	nodes := make([]*protos.DiagramNode, len(d.Nodes))
	for i, n := range d.Nodes {
		nodes[i] = toProtoDiagramNode(n)
	}

	edges := make([]*protos.DiagramEdge, len(d.Edges))
	for i, e := range d.Edges {
		edges[i] = toProtoDiagramEdge(e)
	}

	return &protos.SystemDiagram{
		SystemName: d.SystemName,
		Nodes:      nodes,
		Edges:      edges,
	}
}

func fromProtoSystemDiagram(p *protos.SystemDiagram) *SystemDiagram {
	if p == nil {
		return nil
	}

	nodes := make([]DiagramNode, len(p.Nodes))
	for i, n := range p.Nodes {
		nodes[i] = fromProtoDiagramNode(n)
	}

	edges := make([]DiagramEdge, len(p.Edges))
	for i, e := range p.Edges {
		edges[i] = fromProtoDiagramEdge(e)
	}

	return &SystemDiagram{
		SystemName: p.SystemName,
		Nodes:      nodes,
		Edges:      edges,
	}
}

func toProtoDiagramNode(n DiagramNode) *protos.DiagramNode {
	methods := make([]*protos.MethodInfo, len(n.Methods))
	for i, m := range n.Methods {
		methods[i] = &protos.MethodInfo{
			Name:       m.Name,
			ReturnType: m.ReturnType,
			Traffic:    m.Traffic,
		}
	}

	return &protos.DiagramNode{
		Id:       n.ID,
		Name:     n.Name,
		Type:     n.Type,
		Methods:  methods,
		Traffic:  n.Traffic,
		FullPath: n.FullPath,
		Icon:     n.Icon,
	}
}

func fromProtoDiagramNode(p *protos.DiagramNode) DiagramNode {
	if p == nil {
		return DiagramNode{}
	}

	methods := make([]MethodInfo, len(p.Methods))
	for i, m := range p.Methods {
		methods[i] = MethodInfo{
			Name:       m.Name,
			ReturnType: m.ReturnType,
			Traffic:    m.Traffic,
		}
	}

	return DiagramNode{
		ID:       p.Id,
		Name:     p.Name,
		Type:     p.Type,
		Methods:  methods,
		Traffic:  p.Traffic,
		FullPath: p.FullPath,
		Icon:     p.Icon,
	}
}

func toProtoDiagramEdge(e DiagramEdge) *protos.DiagramEdge {
	return &protos.DiagramEdge{
		FromId:      e.FromID,
		ToId:        e.ToID,
		FromMethod:  e.FromMethod,
		ToMethod:    e.ToMethod,
		Label:       e.Label,
		Order:       e.Order,
		Condition:   e.Condition,
		Probability: e.Probability,
		GeneratorId: e.GeneratorID,
		Color:       e.Color,
	}
}

func fromProtoDiagramEdge(p *protos.DiagramEdge) DiagramEdge {
	if p == nil {
		return DiagramEdge{}
	}

	return DiagramEdge{
		FromID:      p.FromId,
		ToID:        p.ToId,
		FromMethod:  p.FromMethod,
		ToMethod:    p.ToMethod,
		Label:       p.Label,
		Order:       p.Order,
		Condition:   p.Condition,
		Probability: p.Probability,
		GeneratorID: p.GeneratorId,
		Color:       p.Color,
	}
}

// Canvas state conversions

func toProtoCanvas(c *CanvasState) *protos.Canvas {
	if c == nil {
		return nil
	}

	generators := make([]*protos.Generator, len(c.Generators))
	for i, g := range c.Generators {
		generators[i] = toProtoGenerator(&g)
	}

	metrics := make([]*protos.Metric, len(c.Metrics))
	for i, m := range c.Metrics {
		metrics[i] = toProtoMetric(&m)
	}

	return &protos.Canvas{
		CreatedAt:    timestamppb.New(c.CreatedAt),
		UpdatedAt:    timestamppb.New(c.UpdatedAt),
		Id:           c.ID,
		ActiveSystem: c.ActiveSystem,
		LoadedFiles:  c.LoadedFiles,
		Generators:   generators,
		Metrics:      metrics,
	}
}

func fromProtoCanvas(p *protos.Canvas) *CanvasState {
	if p == nil {
		return nil
	}

	generators := make([]Generator, len(p.Generators))
	for i, g := range p.Generators {
		if gen := fromProtoGenerator(g); gen != nil {
			generators[i] = *gen
		}
	}

	metrics := make([]Metric, len(p.Metrics))
	for i, m := range p.Metrics {
		if met := fromProtoMetric(m); met != nil {
			metrics[i] = *met
		}
	}

	c := &CanvasState{
		ID:           p.Id,
		ActiveSystem: p.ActiveSystem,
		LoadedFiles:  p.LoadedFiles,
		Generators:   generators,
		Metrics:      metrics,
	}

	if p.CreatedAt != nil {
		c.CreatedAt = p.CreatedAt.AsTime()
	}
	if p.UpdatedAt != nil {
		c.UpdatedAt = p.UpdatedAt.AsTime()
	}

	return c
}

// Helper functions for slices

func toProtoGenerators(gens []*Generator) []*protos.Generator {
	result := make([]*protos.Generator, len(gens))
	for i, g := range gens {
		result[i] = toProtoGenerator(g)
	}
	return result
}

func fromProtoGenerators(protos []*protos.Generator) []*Generator {
	result := make([]*Generator, len(protos))
	for i, p := range protos {
		result[i] = fromProtoGenerator(p)
	}
	return result
}

func toProtoMetrics(mets []*Metric) []*protos.Metric {
	result := make([]*protos.Metric, len(mets))
	for i, m := range mets {
		result[i] = toProtoMetric(m)
	}
	return result
}

func fromProtoMetrics(protos []*protos.Metric) []*Metric {
	result := make([]*Metric, len(protos))
	for i, p := range protos {
		result[i] = fromProtoMetric(p)
	}
	return result
}
