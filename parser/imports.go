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
type WaitStmt = decl.WaitStmt
type LogStmt = decl.LogStmt
type DelayStmt = decl.DelayStmt
type TypeDecl = decl.TypeDecl
type ParamDecl = decl.ParamDecl
type ComponentDecl = decl.ComponentDecl
type SystemDecl = decl.SystemDecl
type EnumDecl = decl.EnumDecl
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
type GoStmt = decl.GoStmt
type AssignmentStmt = decl.AssignmentStmt
type OptionsDecl = decl.OptionsDecl
type ImportDecl = decl.ImportDecl

// Slices for lists
type ComponentDeclBodyItem = decl.ComponentDeclBodyItem
type SystemDeclBodyItem = decl.SystemDeclBodyItem

type BinaryExpr = decl.BinaryExpr
type UnaryExpr = decl.UnaryExpr
type MemberAccessExpr = decl.MemberAccessExpr
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
