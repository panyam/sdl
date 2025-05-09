%{
package parser

import (
    // "reflect"
    "log"
    "fmt"
    "io"
	  gfn "github.com/panyam/goutils/fn"
)

// Function to be called by yyParse on error.
// Needs access to the lexer passed via %parse-param.
func yyerror(yyl yyLexer, msg string) {
    lexer := yyl.(LexerInterface)
	  // line, col := lexer.Position()
    // log.Println("YYERROR MSG = ", msg)
	  // errMsg := fmt.Sprintf("Error at Line %d, Col %d, Near ('%s'): %s", line, col, /* tokenString(lexer.LastToken()),*/ lexer.Text(), msg) // Added tokenString helper call
    lexer.Error(msg)
}

func yyerrok(lexer yyLexer) {
  log.Println("Error here... not sure how to recover")
  ///ErrFlag = 0 
}

%}

// Add %parse-param to accept the lexer instance

%union {
    // Basic types from lexer
    sval string // Holds raw string values like identifiers, literal content

    // AST Nodes (using pointers) - these should have NodeInfo
    file        *File
    componentDecl *ComponentDecl
    systemDecl *SystemDecl
    node        Node // Generic interface for lists and for accessing NodeInfo
    // tokenNode   TokenNode // Generic interface for lists and for accessing NodeInfo
    expr        Expr
    stmt        Stmt
    typeName    *TypeName
    paramDecl   *ParamDecl
    usesDecl    *UsesDecl
    methodDef   *MethodDecl
    instanceDecl *InstanceDecl
    analyzeDecl *AnalyzeDecl
    expectBlock *ExpectationsDecl
    expectStmt  *ExpectStmt
    blockStmt   *BlockStmt
    ifStmt      *IfStmt
    distributeStmt *DistributeStmt
    distributeCase *DistributeCase
    distributeExpr *DistributeExpr
    distributeExprCase *DistributeExprCase
    defaultCase    *DefaultCase
    goStmt         *GoStmt
    assignStmt     *AssignmentStmt
    optionsDecl    *OptionsDecl
    enumDecl       *EnumDecl
    importDecl     *ImportDecl
    waitStmt *WaitStmt
    delayStmt *DelayStmt

    // Slices for lists
    nodeList           []Node
    importDeclList     []*ImportDecl
    compBodyItem        ComponentDeclBodyItem
    compBodyItemList   []ComponentDeclBodyItem
    sysBodyItemList    []SystemDeclBodyItem
    paramList          []*ParamDecl
    assignList         []*AssignmentStmt
    exprList           []Expr
    stmtList           []Stmt
    ident              *IdentifierExpr
    identList          []*IdentifierExpr
    distributeCaseList []*DistributeCase
    distributeExprCaseList []*DistributeExprCase
    expectStmtList     []*ExpectStmt
    methodSigItemList []*MethodDecl

    // Add field to store position for simple tokens if needed
    // posInfo     NodeInfo
}

// --- Tokens ---
// Keywords (assume lexer returns token type, parser might need pos for some)
%token<node> SYSTEM USES METHOD INSTANCE ANALYZE EXPECT LET IF ELSE DISTRIBUTE DEFAULT RETURN DELAY WAIT GO LOG SWITCH CASE FOR 

// Marking these as nodes so can be returned as Node for their locations
%token<node> USE NATIVE LBRACE RBRACE OPTIONS ENUM COMPONENT PARAM IMPORT FROM AS

// Operators and Punctuation (assume lexer returns token type, use $N.(Node).Pos() if $N is a literal/ident)
%token<node> ASSIGN COLON LPAREN RPAREN COMMA DOT ARROW PLUS_ASSIGN MINUS_ASSIGN MUL_ASSIGN DIV_ASSIGN LET_ASSIGN  SEMICOLON 

%token<node>  INT FLOAT BOOL STRING DURATION NOT MINUS 

// Literals (lexer provides *LiteralExpr or *IdentifierExpr in lval.expr, with NodeInfo)
%token <expr> INT_LITERAL FLOAT_LITERAL STRING_LITERAL BOOL_LITERAL DURATION_LITERAL
%token <ident> IDENTIFIER

// Operators (Tokens for precedence rules, lexer provides string in lval.sval)
%token <node> OR AND EQ NEQ LT LTE GT GTE PLUS MUL DIV MOD

// --- Types (Associating non-terminals with union fields) ---
%type <file>         File
%type <nodeList>     DeclarationList
%type <node>         SystemBodyItem  TopLevelDeclaration
%type <componentDecl>         ComponentDecl
%type <systemDecl>         SystemDecl
%type <compBodyItem> ComponentBodyItem 
%type <compBodyItemList> ComponentBodyItemList ComponentBodyItemOptList
%type <sysBodyItemList>  SystemBodyItemOptList 
%type <optionsDecl>  OptionsDecl
%type <enumDecl>     EnumDecl
%type <identList>    IdentifierList CommaIdentifierListOpt
%type <importDecl>   ImportDecl ImportItem
%type <importDeclList>   ImportList ImportListOpt
%type <stmt>         Stmt IfStmtElseOpt LetStmt ExprStmt ReturnStmt LogStmt SwitchStmt
%type <waitStmt>     WaitStmt
%type <delayStmt>    DelayStmt 
%type <assignStmt>   AssignStmt
%type <blockStmt>    BlockStmt
%type <stmtList>     StmtList 
%type <expr>         Expression OrExpr AndBoolExpr CmpExpr AddExpr MulExpr UnaryExpr PrimaryExpr LiteralExpr CallExpr MemberAccessExpr SwitchExpr CaseExpr DefaultCaseExpr // Added SwitchExpr, CaseExpr, DefaultCaseExpr
%type <exprList>     ArgList ArgListOpt CommaExpressionListOpt
%type <paramDecl>    ParamDecl MethodParamDecl
%type <paramList>    MethodParamList MethodParamListOpt
%type <typeName>     TypeName PrimitiveType // FUTURE USE: OutcomeType QualifiedIdentifier
%type <usesDecl>     UsesDecl
%type <methodDef>    MethodDecl
%type <methodDef>    MethodSigDecl
%type <instanceDecl> InstanceDecl
%type <assignStmt>   Assignment
%type <assignList>   AssignList  AssignListOpt
%type <analyzeDecl>  AnalyzeDecl
%type <expectBlock>  ExpectBlock ExpectBlockOpt
%type <expectStmt>   ExpectStmt
%type <expectStmtList> ExpectStmtList ExpectStmtOptList
%type <methodSigItemList> MethodSigDeclList MethodSigDeclOptList
%type <ifStmt>       IfStmt
%type <distributeStmt> DistributeStmt 
%type <expr>          TotalClauseOpt 
%type <distributeCase> DistributeCase 
%type <distributeCaseList> DistributeCaseListOpt 
%type <defaultCase>    DefaultCase DefaultCaseOpt
%type <distributeExpr> DistributeExpr 
%type <distributeExprCaseList> DistributeExprCaseListOpt 
%type <expr>           DefaultCaseExpr DefaultCaseExprOpt // Expr for cases
%type <distributeExprCase>           DistributeExprCase
%type <goStmt>         GoStmt

// --- Operator Precedence and Associativity (Lowest to Highest) ---
%left OR
%left AND
%nonassoc EQ NEQ LT LTE GT GTE // Comparisons don't chain
%left PLUS MINUS
%left MUL DIV MOD
%right NOT UMINUS // Unary operators (UMINUS for precedence)

%%
// --- Grammar Rules (with position info derived from $N) ---

File:
    DeclarationList {
      ni := NodeInfo{}
      if len($1) > 0 {
        ni.StartPos = $1[0].Pos()
        ni.StopPos = $1[len($1)-1].End()
      }
      yylex.(*Lexer).parseResult = &File{NodeInfo: ni, Declarations: $1}
      // $$ = &File{NodeInfo: ni, Declarations: $1}
    } 
    ;

DeclarationList:
    /* empty */         { $$ = []Node{} }
    | DeclarationList TopLevelDeclaration { $$ = append($1, $2) }
    ;

TopLevelDeclaration:
      ComponentDecl { $$ = $1 }
    | SystemDecl    { $$ = $1 }
    | OptionsDecl   { $$ = $1 }
    | EnumDecl      { $$ = $1 }
    | ImportDecl    { $$ = $1 }
    ;

OptionsDecl:
    OPTIONS LBRACE StmtList RBRACE { // OPTIONS ($1) LBRACE ($2) StmtList ($3) RBRACE ($4)
        // Assume OPTIONS token itself doesn't carry complex NodeInfo from lexer for this example.
        // Span from LBRACE to RBRACE for body. If StmtList is empty, Body.NodeInfo might be tricky.
/*
        bodyStart := $2.(Node).Pos() // Position of LBRACE (assuming lexer returns it as Node)
        bodyEnd := $4.(Node).Pos()   // Position of RBRACE (actually its start, use .End() for full span)
        if len($3) > 0 { // If StmtList is not empty
            bodyStart = $3[0].Pos()
            bodyEnd = $3[len($3)-1].End()
        }
*/
        $$ = &OptionsDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $4.(Node).End()), // Pos of OPTIONS, End of RBRACE
            Body: &BlockStmt{
                 NodeInfo: newNodeInfo($2.(Node).Pos(), $4.(Node).End()),
                 Statements: $3,
             },
         }
    }
    ;

ComponentDecl:
    NATIVE COMPONENT IDENTIFIER LBRACE MethodSigDeclOptList RBRACE { // COMPONENT($1) ... RBRACE($5)
        $$ = &ComponentDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $6.(Node).End()),
            NameNode: $3,
            Body: gfn.Map($5, func(m *MethodDecl) ComponentDeclBodyItem { return m }),
         }
    }
    | COMPONENT IDENTIFIER LBRACE ComponentBodyItemOptList RBRACE { // COMPONENT($1) ... RBRACE($5)
        $$ = &ComponentDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $5.(Node).End()),
            NameNode: $2,
            Body: $4,
         }
    }
    ;

EnumDecl:
    ENUM IDENTIFIER LBRACE IdentifierList RBRACE { // ENUM($1) IDENTIFIER($2) ... RBRACE($5)
        $$ = &EnumDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $5.(Node).End()),
            NameNode: $2, // $2 is an IdentifierExpr from lexer, has Pos/End
            ValuesNode: $4,
        }
    }
    ;

IdentifierList:
    IDENTIFIER                { $$ = []*IdentifierExpr{$1} }
    | IdentifierList COMMA IDENTIFIER { $$ = append($1, $3) }
    ;

ImportDecl:
    IMPORT STRING_LITERAL { // IMPORT($1) STRING_LITERAL($2)
        $$ = &ImportDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $2.End()),
            Path: $2.(*LiteralExpr), // $2 is a LiteralExpr from lexer
        }
    }
    ;

ImportListOpt: /* empty */ { $$ = nil }
             | ImportList { $$ = $1 }
            ;

ImportList : ImportItem             { $$ = []*ImportDecl{$1}; }
           | ImportList ImportItem  { $$ = append($$, $1...) }
           ;

ImportItem: IDENTIFIER { $$ = &ImportDecl{ImportedItem: $1} }
          | IDENTIFIER AS IDENTIFIER { $$ = &ImportDecl{ImportedItem: $1, Alias: $3 } }
          ;

MethodSigDeclOptList:
                /* empty */ { $$ = []*MethodDecl{} }
              | MethodSigDeclList { $$ = $1 }
              ;

MethodSigDeclList:
              MethodSigDecl { $$=[]*MethodDecl{$1} }
              | MethodSigDeclList MethodSigDecl { $$=append($1, $2) };

MethodSigDecl:
    METHOD IDENTIFIER LPAREN MethodParamListOpt RPAREN { // METHOD($1) ... BlockStmt($6)
        $$ = &MethodDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $5.End()),
            NameNode: $2,
            Parameters: $4,
        }
    }
    | METHOD IDENTIFIER LPAREN MethodParamListOpt RPAREN TypeName { // METHOD($1) ... BlockStmt($8)
        $$ = &MethodDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $6.End()),
            NameNode: $2,
            Parameters: $4,
            ReturnType: $6,
         }
    }
    ;

ComponentBodyItemOptList:
                /* empty */ { $$ = []ComponentDeclBodyItem{} }
              | ComponentBodyItemList { $$ = $1 }
              ;

ComponentBodyItemList:
                ComponentBodyItem { $$=[]ComponentDeclBodyItem{$1} }
              | ComponentBodyItemList ComponentBodyItem { $$=append($1, $2) }
              ;

ComponentBodyItem:
      ParamDecl   { $$ = $1 }
    | UsesDecl    { $$ = $1 }
    | MethodDecl   { $$ = $1 }
    | ComponentDecl { $$ = $1 } // Allow nested components
    ;

ParamDecl:
    PARAM IDENTIFIER TypeName { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()),
            Name: $2,
            Type: $3, // TypeName also needs to have NodeInfo
        }
    }
    | PARAM IDENTIFIER TypeName ASSIGN Expression { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $5.End()),
            Name: $2,
            Type: $3,
            DefaultValue: $5,
        }
    }
    ;

TypeName:
      // PrimitiveType { $$ = $1 } // PrimitiveType actions set NodeInfo
    IDENTIFIER    {
                      identNode := $1
                      $$ = &TypeName{
                           NodeInfo: identNode.NodeInfo,
                           EnumTypeName: identNode.Name,
                      }
                    }
    // | QualifiedIdentifier { $$ = $1 } // For future pkg.Type
    // | OutcomeType { $$ = $1 } // Need separate rule if we allow Outcome[T] syntax
    ;

PrimitiveType: // These are keywords. Assume lexer sets NodeInfo if $N.posInfo is used
      INT      { $$ = &TypeName{NodeInfo: $1.(*TokenNode).NodeInfo, PrimitiveTypeName: "int"} }
    | FLOAT    { $$ = &TypeName{NodeInfo: $1.(*TokenNode).NodeInfo, PrimitiveTypeName: "float"} }
    | STRING   { $$ = &TypeName{NodeInfo: $1.(*TokenNode).NodeInfo, PrimitiveTypeName: "string"} }
    | BOOL     { $$ = &TypeName{NodeInfo: $1.(*TokenNode).NodeInfo, PrimitiveTypeName: "bool"} }
    | DURATION { $$ = &TypeName{NodeInfo: $1.(*TokenNode).NodeInfo, PrimitiveTypeName: "duration"} }
    ;

// Placeholder for future pkg.Type or Outcome[T]
// QualifiedIdentifier: IDENTIFIER DOT IDENTIFIER { ... }
// OutcomeType: "Outcome" LBRACKET PrimitiveType RBRACKET { ... }

UsesDecl:
    USES IDENTIFIER IDENTIFIER { // USES($1) ... 
        $$ = &UsesDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()),
            NameNode: $2,
            ComponentNode: $3,
         }
    }
    // Optional: Add syntax for overrides within uses? Like `uses x: T { p1 = v1; }`
    // USES IDENTIFIER IDENTIFIER LBRACE AssignListOpt RBRACE { ... }
    ;

MethodDecl:
    METHOD IDENTIFIER LPAREN MethodParamListOpt RPAREN BlockStmt { // METHOD($1) ... BlockStmt($6)
        $$ = &MethodDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $6.End()),
            NameNode: $2,
            Parameters: $4,
            Body: $6,
        }
    }
    | METHOD IDENTIFIER LPAREN MethodParamListOpt RPAREN TypeName BlockStmt { // METHOD($1) ... BlockStmt($8)
        $$ = &MethodDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $7.End()),
            NameNode: $2,
            Parameters: $4,
            ReturnType: $6,
            Body: $7,
         }
    }
    ;

MethodParamListOpt:
    /* empty */ { $$ = []*ParamDecl{} }
    | MethodParamList { $$ = $1 }
    ;

MethodParamList:
    MethodParamDecl               { $$ = []*ParamDecl{$1} }
    | MethodParamList COMMA MethodParamDecl { $$ = append($1, $3) }
    ;

MethodParamDecl:    // thse dont need "param" unlike param decls in components
    IDENTIFIER TypeName { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: newNodeInfo($1.Pos(), $2.End()),
            Name: $1,
            Type: $2, // TypeName also needs to have NodeInfo
        }
    }
    | IDENTIFIER TypeName ASSIGN Expression { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: newNodeInfo($1.Pos(), $4.End()),
            Name: $1,
            Type: $2,
            DefaultValue: $4,
        }
    }
    ;

// --- System ---
SystemDecl:
    SYSTEM IDENTIFIER LBRACE SystemBodyItemOptList RBRACE { // SYSTEM($1) ... RBRACE($5)
        $$ = &SystemDecl{
             NodeInfo: newNodeInfo($1.(Node).Pos(), $5.(Node).End()),
             NameNode: $2,
             Body: $4,
        }
    }
    ;

SystemBodyItemOptList:
    /* empty */             { $$ = []SystemDeclBodyItem{} }
    | SystemBodyItemOptList SystemBodyItem { $$ = append($1, $2.(SystemDeclBodyItem)) }
    ;

SystemBodyItem:
              InstanceDecl { $$=$1 }
            // | AnalyzeDecl { $$=$1 }
            | OptionsDecl { $$=$1 }
            | LetStmt { $$=$1 }
            ;

InstanceDecl:
    USE IDENTIFIER IDENTIFIER { // IDENTIFIER($1) ... 
         $$ = &InstanceDecl{
             NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()),
             NameNode: $2,
             ComponentType: $3,
             Overrides: []*AssignmentStmt{},
         }
    }
    | USE IDENTIFIER IDENTIFIER ASSIGN LBRACE AssignListOpt RBRACE { // IDENTIFIER($1) ... 
        $$ = &InstanceDecl{
             NodeInfo: newNodeInfo($1.(Node).Pos(), $7.End()),
             NameNode: $2,
             ComponentType: $3,
             Overrides: $6,
         }
    }
    ;

AssignListOpt:
      /* empty */  { $$ = []*AssignmentStmt{} }
    | AssignList { $$ = $1 }
    ;

AssignList:
    Assignment             { $$ = []*AssignmentStmt{$1} }
    | AssignList Assignment { $$ = append($1, $2) }
    ;

Assignment:
    IDENTIFIER ASSIGN Expression { // IDENTIFIER($1) ... 
        $$ = &AssignmentStmt{
            NodeInfo: newNodeInfo($1.Pos(), $3.End()),
            Var: $1,
            Value: $3,
         }
    }
    ;

AnalyzeDecl:
    ANALYZE IDENTIFIER ASSIGN Expression ExpectBlockOpt { // ANALYZE($1) ... 
        callExpr, ok := $4.(*CallExpr)
        if !ok {
            yyerror(yylex, fmt.Sprintf("analyze target must be a method call, found %T at pos %d", $4, $4.(Node).Pos()))
        }
        endPos := 0
        if $5 != nil {
          endPos = $5.End()
        } else {
          endPos = $4.End()
        }
        $$ = &AnalyzeDecl{
             NodeInfo: newNodeInfo($1.(Node).Pos(), endPos),
             Name: $2,
             Target: callExpr,
             Expectations: $5,
         }
    }
    ;

ExpectBlockOpt: /* empty */   { $$ = nil } | ExpectBlock { $$ = $1 };

ExpectBlock:
    EXPECT LBRACE ExpectStmtOptList RBRACE { // EXPECT($1) ... RBRACE($4)
        log.Println("Did Expect Block Hit?")
        $$ = &ExpectationsDecl{
            NodeInfo: newNodeInfo($1.(Node).Pos(), $4.(Node).End()),
            Expects: $3,
        }
    }
    ;

ExpectStmtOptList:
    /* empty */ { $$ = []*ExpectStmt{} }
    | ExpectStmtList { $$ = $1 }
    ;

ExpectStmtList:
    ExpectStmt {
      log.Println("Did we come here????")
      $$ = []*ExpectStmt{$1}
    }
    | ExpectStmtList SEMICOLON CmpExpr {
      log.Println("Why not here Did we come here????")
        cmpExp := $3.(*BinaryExpr);
        expct := &ExpectStmt{ NodeInfo: newNodeInfo($1[0].Pos(), $3.End()), Target: cmpExp.Left.(*MemberAccessExpr), Operator: cmpExp.Operator, Threshold: cmpExp.Right}
        $$ = append($1, expct)
    }
    ;

ExpectStmt:
    Expression EQ Expression { $$ = &ExpectStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: "==", Threshold: $3} }
    | Expression NEQ Expression { $$ = &ExpectStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: "!=", Threshold: $3} }
    | Expression LT Expression { $$ = &ExpectStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: "<", Threshold: $3} }
    | Expression LTE Expression { $$ = &ExpectStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: "<=", Threshold: $3} }
    | Expression GT Expression { $$ = &ExpectStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: ">", Threshold: $3} }
    | Expression GTE Expression { $$ = &ExpectStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: ">=", Threshold: $3} }
    ;

// --- Statements ---
StmtList: 
            /* empty */       { $$ = []Stmt{} }
        | StmtList Stmt { $$ = append($1, $2) } ;

Stmt:
      LetStmt        { $$ = $1 }
    | ExprStmt       { $$ = $1 }
    | ReturnStmt     { $$ = $1 }
    | IfStmt         { $$ = $1 }
    | WaitStmt       { $$ = $1 }
    | DelayStmt      { $$ = $1 }
    | GoStmt         { $$ = $1 }
    | LogStmt        { $$ = $1 }
    | BlockStmt      { $$ = $1 }
    | DistributeStmt { $$ = $1 }
    | SwitchStmt     { $$ = $1 } // Add SwitchStmt
    // | AssignStmt     { $$ = $1 } // Disallow assignments as statements? Let's require `let` for now.
    | error SEMICOLON { yyerrok(yylex) /* Recover on semicolon */ } // Basic error recovery
    ;


BlockStmt:
    LBRACE StmtList RBRACE {
      $$ = &BlockStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Statements: $2 }
    }
    ;

LetStmt:
    LET IDENTIFIER ASSIGN Expression { // LET($1) ... 
         $$ = &LetStmt{
             NodeInfo: newNodeInfo($1.(Node).Pos(), $4.End()),
             Variable: $2,
             Value: $4,
          }
    }
    ;

AssignStmt: // Rule for simple assignment `a = b;` if needed as statement
    IDENTIFIER ASSIGN Expression {
         // This might conflict with InstanceDecl's Assignment rule if not careful.
         // Let's prefer LetStmt for variables. This rule might be removed.
         // For now, map it to AssignmentStmt AST node used by InstanceDecl.
         $$ = &AssignmentStmt{
             NodeInfo: newNodeInfo($1.Pos(), $3.Pos()),
             Var: $1,
             Value: $3,
         }
    }
    ;

ExprStmt:
    Expression SEMICOLON { $$ = &ExprStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $2.(Node).End()), Expression: $1 } }
    ;

ReturnStmt:
    RETURN Expression { $$ = &ReturnStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $2.(Node).End()), ReturnValue: $2 } }
    | RETURN SEMICOLON          { $$ = &ReturnStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $2.(Node).End()), ReturnValue: nil } }
    ;

DelayStmt:
    DELAY Expression { $$ = &DelayStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $2.End()), Duration: $2 } }
    ;

WaitStmt:
    WAIT IdentifierList { // WAIT($1) IDENTIFIER($2) ... 
         idents := $2
         endNode := idents[len(idents)-1] // End at the last identifier in the list
         $$ = &WaitStmt{ NodeInfo: newNodeInfo($1.Pos(), endNode.End()), Idents: idents }
    }
    ;

CommaIdentifierListOpt:
      /* empty */                  { $$ = nil }
    | COMMA IDENTIFIER CommaIdentifierListOpt {
         $$ = []*IdentifierExpr{$2}
         if $3 != nil { $$ = append($$, $3...) }
    }
    ;

LogStmt:
    LOG Expression CommaExpressionListOpt { // LOG($1) Expression($2) ... 
        args := []Expr{$2}
        endPos := 0
        exprList := $3
        if exprList != nil {
          endPos = exprList[len(exprList) - 1].End()
        } else {
          endPos = $2.End()
        }
        if len(exprList) > 0 { // $3 is CommaExpressionListOpt -> []Expr
            args = append(args, $3...)
        }
         $$ = &LogStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), endPos), Args: args }
    }
    ;

CommaExpressionListOpt:
      /* empty */                   { $$ = nil }
    | COMMA Expression CommaExpressionListOpt {
        $$ = []Expr{$2}
        if $3 != nil { $$ = append($$, $3...) }
    }
    ;


// --- Control Flow ---
IfStmt:
    IF Expression BlockStmt IfStmtElseOpt { // IF($1) ...
        endNode := Stmt($3)
        if $4 != nil { endNode = $4 } // End of Else block/IfStmt
        $$ = &IfStmt{
          NodeInfo: newNodeInfo($1.(Node).Pos(), endNode.End()),
          Condition: $2,
          Then: $3,
          Else: $4,
        }
    }
    ;

IfStmtElseOpt:
      /* empty */         { $$ = nil }
    | ELSE IfStmt         { $$ = $2 } // Chain if-else
    | ELSE BlockStmt      { $$ = $2 } // else { ... }
    ;

DistributeStmt:
    DISTRIBUTE TotalClauseOpt LBRACE DistributeCaseListOpt DefaultCaseOpt RBRACE { // DISTRIBUTE($1) ... RBRACE($6)
        $$ = &DistributeStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $6.(Node).End()), Total: $2, Cases: $4, DefaultCase: $5 }
    }
    ;

TotalClauseOpt: /* empty */ { $$=nil } | Expression { $$=$1 };

DistributeCaseListOpt:
      /* empty */                  { $$ = []*DistributeCase{} }
    | DistributeCaseListOpt DistributeCase { $$ = append($1, $2) }
    ;

DistributeCase:
    Expression ARROW Stmt { $$ = &DistributeCase{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()), Probability: $1, Body: $3 } }
    ;

DefaultCaseOpt:
      /* empty */   { $$ = nil }
    | DefaultCase { $$ = $1 }
    ;

DefaultCase:
    DEFAULT ARROW Stmt { $$ = &DefaultCase{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()), Body: $3 } }
    ;

GoStmt:
    GO IDENTIFIER ASSIGN Stmt { // GO($1) ... BlockStmt($4)
        $$ = &GoStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $4.End()), VarName: $2, Stmt: $4 }
    }
    | GO BlockStmt { // GO($1) BlockStmt($2)
        $$ = &GoStmt{ NodeInfo: newNodeInfo($1.(Node).Pos(), $2.End()), VarName: nil, Stmt: $2 }
    }
    | GO IDENTIFIER ASSIGN Expression {
         yyerror(yylex, fmt.Sprintf("`go` currently only supports assigning blocks, not expressions, at pos %d", $1.(Node).Pos()))
         $$ = &GoStmt{}
    }
    ;

SwitchStmt: // Placeholder
    SWITCH Expression LBRACE /* TODO */ RBRACE { yyerror(yylex, "Switch statement not defined"); $$ = nil };

// --- Expressions ---
Expression: OrExpr { $$ = $1 } ;
OrExpr: AndBoolExpr { $$=$1 }
      | OrExpr OR AndBoolExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
      ;

AndBoolExpr:
       CmpExpr { $$=$1 }
      | AndBoolExpr AND CmpExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.(Node).String(), Right: $3} }
      ;

CmpExpr: AddExpr { $$=$1 }
    | AddExpr EQ AddExpr  { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    | AddExpr NEQ AddExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    | AddExpr LT AddExpr  { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    | AddExpr LTE AddExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    | AddExpr GT AddExpr  { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    | AddExpr GTE AddExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    ;

AddExpr:
      MulExpr { $$=$1 }
    | AddExpr PLUS MulExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    | AddExpr MINUS MulExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    ;

MulExpr: UnaryExpr { $$=$1 }
    | MulExpr MUL UnaryExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    | MulExpr DIV UnaryExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    | MulExpr MOD UnaryExpr { $$ = &BinaryExpr{ NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()), Left: $1, Operator: $2.String(), Right: $3} }
    ;

UnaryExpr: PrimaryExpr { $$=$1 }
    // For Unary, $1 is operator token, $2 is operand Expr node
    | NOT UnaryExpr   { $$ = &UnaryExpr{ NodeInfo: newNodeInfo($1.Pos(), $2.(Node).End()), Operator: $1.String(), Right: $2} }
    | MINUS UnaryExpr %prec UMINUS { $$ = &UnaryExpr{ NodeInfo: newNodeInfo($1.Pos(), $2.(Node).End()), Operator: $1.String(), Right: $2} }
    ;

PrimaryExpr:
      LiteralExpr             { $$ = $1 }
    | IDENTIFIER { $$ = $1 } // Already has position info
    | MemberAccessExpr        { $$ = $1 }
    | CallExpr                { $$ = $1 }
    // | DistributeExpr          { $$ = $1 } // Expression version
    // | SwitchExpr              { $$ = $1 } // Expression version
    | LPAREN Expression RPAREN { $$ = $2 } // Grouping
    ;

LiteralExpr:
      INT_LITERAL             { 
          // yylex.(*Lexer).lval)
          $$ = $1 
      }
    | FLOAT_LITERAL           { $$ = $1 }
    | STRING_LITERAL          { $$ = $1 }
    | BOOL_LITERAL            { $$ = $1 }
    | DURATION_LITERAL        { $$ = $1 }
    ;

MemberAccessExpr:
    PrimaryExpr DOT IDENTIFIER { // PrimaryExpr($1) DOT($2) IDENTIFIER($3)
         $$ = &MemberAccessExpr{
             NodeInfo: newNodeInfo($1.(Node).Pos(), $3.End()),
             Receiver: $1,
             Member: $3,
         }
    }
    ;

CallExpr:
    PrimaryExpr LPAREN RPAREN { // PrimaryExpr($1) LPAREN($2) RPAREN($3)
         $$ = &CallExpr{
             NodeInfo: newNodeInfo($1.(Node).Pos(), $3.(Node).End()),
             Function: $1,
             Args: []Expr{},
         }
    }
    | PrimaryExpr LPAREN ArgList RPAREN { // PrimaryExpr($1) LPAREN($2) ArgList($3) RPAREN($4)
         endNode := $4.(Node) // End at RPAREN
         if len($3) > 0 {
             exprList := $3
             endNode = exprList[len(exprList)-1].(Node) // End at last arg
         }
         $$ = &CallExpr{
             NodeInfo: newNodeInfo($1.(Node).Pos(), endNode.End()),
             Function: $1,
             Args: $3,
         }
    }
    ;

ArgListOpt:
      /* empty */ { $$ = []Expr{} }
    | ArgList   { $$ = $1 }
    ;

ArgList:
    Expression                { $$ = []Expr{$1} }
    | ArgList COMMA Expression { $$ = append($1, $3) }
    ;


DistributeExpr:
    DISTRIBUTE TotalClauseOpt LBRACE DistributeExprCaseListOpt DefaultCaseExprOpt RBRACE {
         $$ = &DistributeExpr{TotalProb: $2, Cases: $4, Default: $5} /* TODO: Pos */
    }
    ;

DistributeExprCaseListOpt:
      /* empty */                         { $$ = []*DistributeExprCase{} }
    | DistributeExprCaseListOpt DistributeExprCase { $$ = append($1, $2) } // Ensure type assertion
    ;

DistributeExprCase: // Returns Expr for use in DistributeExpr struct
    Expression ARROW Expression {
        // Need to wrap in DistributeExprCase AST node
        $$ = &DistributeExprCase{ Probability: $1, Body: $3 } /* TODO: Pos */
    }
    ;

DefaultCaseExprOpt: // Returns Expr
      /* empty */       { $$ = nil }
    | DefaultCaseExpr { $$ = $1 }
    ;

DefaultCaseExpr: // Returns Expr
    DEFAULT ARROW Expression { $$ = $3 } // Return the body expression directly
    ;


SwitchExpr: // Placeholder - Needs AST definition
    SWITCH Expression LBRACE /* Case definitions */ RBRACE {
         yyerror(yylex, "Switch expression not fully defined yet")
         $$ = nil
    }
    ;

CaseExpr: // Placeholder - Needs AST definition
     CASE Expression COLON Expression {
         yyerror(yylex, "Case expression not fully defined yet")
         $$ = nil
     }
     ;
%% // --- Go Code Section ---

// Interface for the lexer required by the parser.
type LexerInterface interface {
    Lex(lval *yySymType) int
    Error(s string)
    Pos() int       // Start byte position of the last token read
    End() int       // End byte position of the last token read
    Text() string   // Text of the last token read
    Position() (line, col int) // Added: Get line/col of last token start
    LastToken() int // Added: Get the token code that was just lexed
}

// Parse takes an input stream and attempts to parse it according to the SDL grammar. 22222
// It returns the root of the Abstract Syntax Tree (*File) if successful, or an error.
func Parse(input io.Reader) (*Lexer, *File, error) {
	// Reset global result before parsing
	lexer := NewLexer(input)
	// Set yyDebug = 3 for verbose parser debugging output
	// yyDebug = 3
	resultCode := yyParse(lexer) // Call the LALR parser generated by goyacc

	if resultCode != 0 {
		// A syntax error occurred. The lexer's Error method should have been called
		// and stored the error message.
		if lexer.lastError != nil {
			return lexer, nil, lexer.lastError
		}
		// Fallback error message if lexer didn't store one
		return lexer, nil, fmt.Errorf("syntax error near byte %d", lexer.Pos())
	}

	// Parsing succeeded
	if lexer.parseResult == nil {
		// This indicates a potential issue with the grammar's top rule action
		return lexer, nil, fmt.Errorf("parsing finished successfully, but no AST result was produced")
	}

	return lexer, lexer.parseResult, nil
}

// The parser expects the lexer variable to be named yyLex.
// We can satisfy this by creating a global or passing it via yyParseWithLexer.
// Using yyParseWithLexer is cleaner.

// Example main function (optional, for standalone testing)
/*
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: parser <input_file>")
		return
	}
	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	astRoot, err := Parse(file)
	if err != nil {
		fmt.Printf("Parsing failed: %v\n", err)
		// Error message should ideally include line/column from lexer
	} else {
		fmt.Println("Parsing successful!")
		// Print the AST (implement String() methods for AST nodes for nice output)
		fmt.Println(astRoot.String())
	}
}
*/
