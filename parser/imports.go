package parser

import "github.com/panyam/leetcoach/sdl/decl"

type NodeInfo = decl.NodeInfo
type LiteralExpr = decl.LiteralExpr
type IdentifierExpr = decl.IdentifierExpr
type Expr = decl.Expr
type Stmt = decl.Stmt
type LetStmt = decl.LetStmt
type ReturnStmt = decl.ReturnStmt
type ExprStmt = decl.ExprStmt
type WaitStmt = decl.WaitStmt
type LogStmt = decl.LogStmt
type DelayStmt = decl.DelayStmt
type TypeName = decl.TypeName
type ParamDecl = decl.ParamDecl
type ComponentDecl = decl.ComponentDecl
type SystemDecl = decl.SystemDecl
type EnumDecl = decl.EnumDecl
type Value = decl.Value

type Node = decl.Node
type File = decl.FileDecl

type ValueType = decl.ValueType
type UsesDecl = decl.UsesDecl
type MethodDecl = decl.MethodDecl
type InstanceDecl = decl.InstanceDecl
type AnalyzeDecl = decl.AnalyzeDecl
type ExpectationsDecl = decl.ExpectationsDecl
type ExpectStmt = decl.ExpectStmt
type BlockStmt = decl.BlockStmt
type IfStmt = decl.IfStmt
type DistributeStmt = decl.DistributeStmt
type DistributeCase = decl.DistributeCase
type DefaultCase = decl.DefaultCase
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
type DistributeExpr = decl.DistributeExpr
type DistributeExprCase = decl.DistributeExprCase

var NewRuntimeValue = decl.NewRuntimeValue
var BoolType = decl.BoolType
var StrType = decl.StrType
var IntType = decl.IntType
var FloatType = decl.FloatType

var BoolValue = decl.BoolValue
var StringValue = decl.StringValue
var IntValue = decl.IntValue
var FloatValue = decl.FloatValue
