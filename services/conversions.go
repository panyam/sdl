package services

import (
	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// SystemDiagram conversions - kept because diagram types are built dynamically
// from runtime state and need conversion to proto for transport

func ToProtoSystemDiagram(d *SystemDiagram) *protos.SystemDiagram {
	if d == nil {
		return nil
	}

	nodes := make([]*protos.DiagramNode, len(d.Nodes))
	for i, n := range d.Nodes {
		nodes[i] = ToProtoDiagramNode(n)
	}

	edges := make([]*protos.DiagramEdge, len(d.Edges))
	for i, e := range d.Edges {
		edges[i] = ToProtoDiagramEdge(e)
	}

	return &protos.SystemDiagram{
		SystemName: d.SystemName,
		Nodes:      nodes,
		Edges:      edges,
	}
}

func ToProtoDiagramNode(n DiagramNode) *protos.DiagramNode {
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

func ToProtoDiagramEdge(e DiagramEdge) *protos.DiagramEdge {
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
