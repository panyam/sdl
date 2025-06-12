package runtime

import (
	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
)

type NodeInfo = decl.NodeInfo
type LiteralExpr = decl.LiteralExpr
type IdentifierExpr = decl.IdentifierExpr
type Location = decl.Location
type Expr = decl.Expr
type Env[T any] = decl.Env[T]
type ExprBase = decl.ExprBase
type Stmt = decl.Stmt
type LetStmt = decl.LetStmt
type SetStmt = decl.SetStmt
type ForStmt = decl.ForStmt
type ReturnStmt = decl.ReturnStmt
type ExprStmt = decl.ExprStmt
type TypeDecl = decl.TypeDecl
type ParamDecl = decl.ParamDecl
type ComponentDecl = decl.ComponentDecl
type SystemDecl = decl.SystemDecl
type EnumDecl = decl.EnumDecl
type Value = decl.Value
type FutureValue = decl.FutureValue
type ThunkValue = decl.ThunkValue
type MethodValue = decl.MethodValue
type RefValue = decl.RefValue

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
type CallExpr = decl.CallExpr
type TupleExpr = decl.TupleExpr
type SampleExpr = decl.SampleExpr

var NewValue = decl.NewValue
var BoolType = decl.BoolType
var StrType = decl.StrType
var IntType = decl.IntType
var FloatType = decl.FloatType

var BoolValue = decl.BoolValue
var StringValue = decl.StringValue
var IntValue = decl.IntValue
var FloatValue = decl.FloatValue
var TupleValue = decl.TupleValue
var Nil = decl.Nil

type ErrorCollector = loader.ErrorCollector

var NewNewExpr = decl.NewNewExpr

var TypeTagFuture = decl.TypeTagFuture
