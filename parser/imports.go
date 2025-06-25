package parser

import "github.com/panyam/sdl/decl"

type NodeInfo = decl.NodeInfo
type LiteralExpr = decl.LiteralExpr
type IdentifierExpr = decl.IdentifierExpr
type Location = decl.Location
type Expr = decl.Expr
type ExprBase = decl.ExprBase
type Stmt = decl.Stmt
type LetStmt = decl.LetStmt
type ForStmt = decl.ForStmt
type ReturnStmt = decl.ReturnStmt
type ExprStmt = decl.ExprStmt
type TypeDecl = decl.TypeDecl
type ParamDecl = decl.ParamDecl
type ComponentDecl = decl.ComponentDecl
type SystemDecl = decl.SystemDecl
type AggregatorDecl = decl.AggregatorDecl
type EnumDecl = decl.EnumDecl
type Annotatable = decl.Annotatable
type AnnotationDecl = decl.AnnotationDecl
type Value = decl.Value

type Node = decl.Node
type FileDecl = decl.FileDecl

type Type = decl.Type
type UsesDecl = decl.UsesDecl
type MethodDecl = decl.MethodDecl
type InstanceDecl = decl.InstanceDecl
type AnalyzeDecl = decl.AnalyzeDecl
type ExpectationsDecl = decl.ExpectationsDecl
type ExpectStmt = decl.ExpectStmt
type BlockStmt = decl.BlockStmt
type IfStmt = decl.IfStmt
type DistributeExpr = decl.DistributeExpr
type CaseExpr = decl.CaseExpr
type CaseStmt = decl.CaseStmt
type SwitchStmt = decl.SwitchStmt
type GoExpr = decl.GoExpr
type WaitExpr = decl.WaitExpr
type AssignmentStmt = decl.AssignmentStmt
type OptionsDecl = decl.OptionsDecl
type ImportDecl = decl.ImportDecl

// Slices for lists
type ComponentDeclBodyItem = decl.ComponentDeclBodyItem
type SystemDeclBodyItem = decl.SystemDeclBodyItem

type BinaryExpr = decl.BinaryExpr
type UnaryExpr = decl.UnaryExpr
type MemberAccessExpr = decl.MemberAccessExpr
type IndexExpr = decl.IndexExpr
type CallExpr = decl.CallExpr
type TupleExpr = decl.TupleExpr
type SampleExpr = decl.SampleExpr

var BoolType = decl.BoolType
var StrType = decl.StrType
var IntType = decl.IntType
var FloatType = decl.FloatType

var BoolValue = decl.BoolValue
var StringValue = decl.StringValue
var IntValue = decl.IntValue
var FloatValue = decl.FloatValue
var NewValue = decl.NewValue
var Nil = decl.Nil

// Helpers for creating nodes
var NewNodeInfo = decl.NewNodeInfo
var NewIntLit = decl.NewIntLit
var NewBoolLit = decl.NewBoolLit
var NewStringLit = decl.NewStringLit
var NewIdent = decl.NewIdent
var NewIdentExpr = decl.NewIdentExpr
var NewLetStmt = decl.NewLetStmt
var NewBinExpr = decl.NewBinExpr
var NewUnaryExpr = decl.NewUnaryExpr
var NewExprStmt = decl.NewExprStmt
var NewReturnStmt = decl.NewReturnStmt
var NewBlockStmt = decl.NewBlockStmt
var NewIfStmt = decl.NewIfStmt
var NewSysDecl = decl.NewSysDecl
var NewInstDecl = decl.NewInstDecl
var NewAssignmentStmt = decl.NewAssignmentStmt
var NewAssignStmt = decl.NewAssignStmt
var NewCompDecl = decl.NewCompDecl
var NewUsesDecl = decl.NewUsesDecl
var NewMethodDecl = decl.NewMethodDecl
var NewTypeDecl = decl.NewTypeDecl
var NewParamDecl = decl.NewParamDecl
var NewMemberAccessExpr = decl.NewMemberAccessExpr
var NewCallExpr = decl.NewCallExpr
var NewNamedCallExpr = decl.NewNamedCallExpr
var NewDistributeExpr = decl.NewDistributeExpr
var NewGoExpr = decl.NewGoExpr
var NewWaitExpr = decl.NewWaitExpr
