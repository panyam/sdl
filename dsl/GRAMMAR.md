// --- Top Level ---
File = Declaration* EOF ;
Declaration = ComponentDecl | SystemDecl | OptionsDecl | ImportDecl /* Future */ | EnumDecl /* NEW */ ;

// *** NEW: Enum Definition ***
EnumDecl = "enum" identifier "{" identifier ("," identifier)* ["..."] "}" ";" ; // Allows defining named discrete value sets

ImportDecl = "import" StringLiteral ";" ;
OptionsDecl = "options" BlockStmt ; // Define structure later


// --- Component Definition ---
ComponentDecl = "component" identifier "{" ComponentBodyItem* "}" ;
ComponentBodyItem = ParamDecl | UsesDecl | OperationDef ; // Removed nested ComponentDecl for now

ParamDecl = "param" identifier ":" TypeName ["=" Expression] ";" ;
// *** Refine TypeName ***
TypeName = PrimitiveType | identifier /* For registered Enums */ ;
PrimitiveType = "int" | "float" | "string" | "bool" | "duration" ;

UsesDecl = "uses" identifier ":" identifier ["{" ParamAssignment* "}"] ";" ; // varName: ComponentType { overrides }

// *** Refine OperationDef ***
OperationDef = "operation" identifier "(" [ParamList] ")" [":" TypeName] BlockStmt ; // TypeName must be primitive or registered Enum
ParamList = ParamDecl ("," ParamDecl)* ;


// --- System Definition ---
SystemDecl = "system" identifier "{" SystemBodyItem* "}" ;
SystemBodyItem = InstanceDecl | AnalyzeDecl | OptionsDecl | LetStmt /* Allow system-level constants? */ ;

InstanceDecl = identifier ":" identifier ["=" "{" ParamAssignment* "}"] ";" ; // instanceName: ComponentType = { overrides }
ParamAssignment = identifier "=" Expression ";" ;

// *** Refine AnalyzeDecl ***
AnalyzeDecl = "analyze" identifier "=" Expression [ExpectBlock] ";" ; // analysisName = instance.Operation(args)
ExpectBlock = "expect" "{" ExpectStmt* "}" ; // Use braces for clarity
ExpectStmt = Expression Operator Expression ";" ; // TargetExpr Op ThresholdExpr (e.g., target.P99 < 100ms)


// --- Statements ---
Stmt = LetStmt | AssignmentStmt | ExprStmt | ReturnStmt | IfStmt
     | DistributeStmt | DelayStmt | RepeatStmt | ParallelStmt | LogStmt
     | BlockStmt /* | SwitchStmt | FilterStmt? */ ;

LetStmt = "let" identifier "=" Expression ";" ;
AssignmentStmt = identifier "=" Expression ";" ; // Maybe disallow, force 'let'? For simplicity, require 'let'. Let's remove this.
// *** Revised: Use LetStmt only ***
Stmt = LetStmt | ExprStmt | ReturnStmt | IfStmt | DistributeStmt | DelayStmt | RepeatStmt | ParallelStmt | LogStmt | BlockStmt ;


ExprStmt = Expression ";" ;
ReturnStmt = "return" Expression ";" ;
DelayStmt = "delay" Expression ";" ; // Expr must yield Duration outcome
LogStmt = "log" (StringLiteral | Expression) ("," Expression)* ";" ;


// --- Control Flow Statements ---
IfStmt = "if" Expression BlockStmt ["else" (IfStmt | BlockStmt)] ; // Condition must yield bool outcome

DistributeStmt = "distribute" [TotalClause] "{" DistributeCase* [DefaultCase] "}" ;
TotalClause = "(" Expression ")" ; // Expr must yield float outcome
DistributeCase = Expression "=>" BlockStmt ; // Expr must yield float outcome
DefaultCase = "default" "=>" BlockStmt ;

RepeatStmt = "repeat" "(" Expression "," ExecutionMode ")" BlockStmt ; // Expr must yield int outcome
ExecutionMode = identifier{"Sequential"} | identifier{"Parallel"} ; // Treat as identifiers for now

ParallelStmt = "parallel" BlockStmt ;


// --- Expressions ---
Expression = OrExpr ; // Precedence: || lowest -> && -> Cmp -> Add/Sub -> Mul/Div -> Unary -> Primary

OrExpr      = AndBoolExpr ( "||" AndBoolExpr )* ;
AndBoolExpr = CmpExpr ( "&&" CmpExpr )* ;
CmpExpr     = AddExpr [ ( "==" | "!=" | "<" | "<=" | ">" | ">=" ) AddExpr ] ;
AddExpr     = MulExpr ( ( "+" | "-" ) MulExpr )* ;
MulExpr     = UnaryExpr ( ( "*" | "/" | "%" ) UnaryExpr )* ;
UnaryExpr   = ( "!" | "-" ) UnaryExpr | PrimaryExpr ;

PrimaryExpr = LiteralExpr
            | IdentifierExpr
            | MemberAccessExpr // instance.param or self.param ONLY
            | CallExpr         // func(args) or instance.method(args)
            //| DistributeExpr   // Keep separate from DistributeStmt for now
            | InternalCallExpr
            | "(" Expression ")" ;


// --- Expression Details ---
LiteralExpr = IntLiteral | FloatLiteral | StringLiteral | BoolLiteral | DurationLiteral ;
DurationLiteral = IntLiteral ("ns" | "us" | "ms" | "s") ; // Allow units

IdentifierExpr = identifier ;

// *** Refine MemberAccessExpr ***
MemberAccessExpr = PrimaryExpr "." identifier ; // Can only access params/fields, not .Success etc.

CallExpr = PrimaryExpr "(" [ArgList] ")" ;
ArgList = Expression ("," Expression)* ;

InternalCallExpr = "Internal" "." identifier "(" [ArgList] ")" ;

// --- Lexical Elements (Simplified) ---
// identifier = letter (letter | digit | "_")* ;
// IntLiteral = digit+ ;
// FloatLiteral = digit+ "." digit+ ;
// StringLiteral = '"' .* '"' ; // Allow escape sequences
// BoolLiteral = "true" | "false" ;
// Comments = "//" ... | "/*" ... "*/" ; (Ignored)
