%{
package parser

import (
    // "reflect"
    "log"
    "fmt"
    "io"
)

// Function to be called by SDLParse on error.
// Needs access to the lexer passed via %parse-param.
func yyerror(yyl SDLLexer, msg string) {
    lexer := yyl.(LexerInterface)
	  // line, col := lexer.Position()
    // log.Println("YYERROR MSG = ", msg)
	  // errMsg := fmt.Sprintf("Error at Line %d, Col %d, Near ('%s'): %s", line, col, /* TokenString(lexer.LastToken()),*/ lexer.Text(), msg) // Added TokenString helper call
    lexer.Error(msg)
}

func yyerrok(lexer SDLLexer) {
  log.Println("Error here... not sure how to recover")
  ///ErrFlag = 0 
}

%}

// Add %parse-param to accept the lexer instance

%union {
    // Basic types from lexer
    sval string // Holds raw string values like identifiers, literal content

    // AST Nodes (using pointers) - these should have NodeInfo
    file        *FileDecl
    componentDecl *ComponentDecl
    systemDecl *SystemDecl
    aggregatorDecl *AggregatorDecl
    node        Node // Generic interface for lists and for accessing NodeInfo
    // tokenNode   TokenNode // Generic interface for lists and for accessing NodeInfo
    expr        Expr
    chainedExpr *ChainedExpr
    stmt        Stmt
    typeDecl    *TypeDecl
    paramDecl   *ParamDecl
    usesDecl    *UsesDecl
    methodDef   *MethodDecl
    instanceDecl *InstanceDecl
    analyzeDecl *AnalyzeDecl
    expectBlock *ExpectationsDecl
    expectStmt  *ExpectStmt
    blockStmt   *BlockStmt
    ifStmt      *IfStmt

    distributeExpr *DistributeExpr
    caseExpr *CaseExpr

    switchStmt *SwitchStmt
    caseStmt *CaseStmt

    tupleExpr *TupleExpr
    goExpr         *GoExpr
    forStmt         *ForStmt
    assignStmt     *AssignmentStmt
    optionsDecl    *OptionsDecl
    enumDecl       *EnumDecl
    importDecl     *ImportDecl
    waitExpr *WaitExpr
    delayStmt *DelayStmt
    sampleExpr *SampleExpr

    // Slices for lists
    nodeList          []Node
    caseExprList      []*CaseExpr
    typeDeclList      []*TypeDecl
    caseStmtList      []*CaseStmt
    importDeclList    []*ImportDecl
    compBodyItem      ComponentDeclBodyItem
    compBodyItemList  []ComponentDeclBodyItem
    sysBodyItemList   []SystemDeclBodyItem
    paramList         []*ParamDecl
    assignList        []*AssignmentStmt
    exprList          []Expr
    stmtList          []Stmt
    ident             *IdentifierExpr
    identList         []*IdentifierExpr
    distributeExprCaseList []*CaseExpr
    expectStmtList     []*ExpectStmt
    methodSigItemList []*MethodDecl

    // Add field to store position for simple tokens if needed
    // posInfo     NodeInfo
}

// --- Tokens ---
// Keywords (assume lexer returns token type, parser might need pos for some)
%token<node> SYSTEM USES AGGREGATOR METHOD ANALYZE EXPECT LET IF ELSE SAMPLE DISTRIBUTE DEFAULT RETURN DELAY WAIT GO GOBATCH USING LOG SWITCH CASE FOR 

// Marking these as nodes so can be returned as Node for their locations
%token<node> USE NATIVE LSQUARE RSQUARE LBRACE RBRACE OPTIONS ENUM COMPONENT PARAM IMPORT FROM AS

// Operators and Punctuation (assume lexer returns token type, use $N.(Node).Pos() if $N is a literal/ident)
%token<node> ASSIGN COLON LPAREN RPAREN COMMA DOT ARROW LET_ASSIGN  SEMICOLON 

%token<node>  INT FLOAT BOOL STRING DURATION

// Literals (lexer provides *LiteralExpr or *IdentifierExpr in lval.expr, with NodeInfo)
%token <expr> INT_LITERAL FLOAT_LITERAL STRING_LITERAL BOOL_LITERAL DURATION_LITERAL
%token <ident> IDENTIFIER

// Operators (Tokens for precedence rules, lexer provides string in lval.sval)
// %token <node> OPERATOR
%token <node> OR AND EQ NEQ LT LTE GT GTE PLUS MUL DIV MOD

%token <node> DUAL_OP BINARY_NC_OP BINARY_OP UNARY_OP MINUS

// --- Types (Associating non-terminals with union fields) ---
%type <file>         File
%type <nodeList>     DeclarationList
%type <node>         SystemBodyItem  TopLevelDeclaration
%type <componentDecl>         ComponentDecl
%type <systemDecl>         SystemDecl
%type <aggregatorDecl>         AggregatorDecl
%type <compBodyItem> ComponentBodyItem 
%type <compBodyItemList> ComponentBodyItemList ComponentBodyItemOptList
%type <compBodyItem> NativeComponentBodyItem 
%type <compBodyItemList> NativeComponentBodyItemList NativeComponentBodyItemOptList
%type <sysBodyItemList>  SystemBodyItemOptList 
%type <optionsDecl>  OptionsDecl
%type <enumDecl>     EnumDecl
%type <identList>    CommaIdentifierList 
%type <importDecl>   ImportItem
%type <importDeclList>   ImportDecl ImportList
%type <stmt>         Stmt IfStmtElseOpt LetStmt ExprStmt ReturnStmt 
%type <delayStmt>    DelayStmt 
%type <sampleExpr>    SampleExpr 
// %type <assignStmt>   AssignStmt
%type <blockStmt>    BlockStmt
%type <stmtList>     StmtList 
%type <tupleExpr>         TupleExpr
%type <expr>         Expression UnaryExpr PrimaryExpr LiteralExpr CallExpr MemberAccessExpr IndexExpr LeafExpr ParenExpr  WaitExpr
// %type <expr>         BinaryExpr NonAssocBinExpr 
%type <chainedExpr>         ChainedExpr
// %type <expr> OrExpr AndBoolExpr CmpExpr AddExpr MulExpr UnaryExpr 
%type <paramDecl>    ParamDecl MethodParamDecl
%type <paramList>    MethodParamList MethodParamListOpt
%type <typeDecl>     TypeDecl
%type <typeDeclList>     TypeDeclList
%type <usesDecl>     UsesDecl
%type <methodDef>    MethodDecl MethodSigDecl
%type <instanceDecl> InstanceDecl
%type <forStmt>   ForStmt
%type <assignStmt>   Assignment
%type <assignList>   AssignList  AssignListOpt
// %type <analyzeDecl>  AnalyzeDecl
// %type <expectBlock>  ExpectBlock ExpectBlockOpt
// %type <expectStmt>   ExpectStmt
// %type <expectStmtList> ExpectStmtList ExpectStmtOptList
// %type <methodSigItemList> MethodSigDeclList MethodSigDeclOptList
%type <ifStmt>       IfStmt
%type <exprList>     CommaSepExprListOpt, CommaSepExprList
// %type <exprList>          OpSepExprList
%type <expr>          TotalClauseOpt 
%type <caseExpr> CaseExpr
%type <caseExprList> CaseExprList CaseExprListOpt
%type <distributeExpr> DistributeExpr 
%type <expr>           DefaultCaseExpr DefaultCaseExprOpt // Expr for cases

%type <switchStmt> SwitchStmt
%type <caseStmt> CaseStmt
%type <caseStmtList> CaseStmtList CaseStmtListOpt
%type <stmt>           DefaultCaseStmt DefaultCaseStmtOpt // Expr for cases

%type <expr>         GoExpr

// --- Operator Precedence and Associativity (Lowest to Highest) ---
%left BINARY_OP MINUS
%nonassoc BINARY_NC_OP 
%left LSQUARE
/*
%left OR
%left AND
%nonassoc EQ NEQ LT LTE GT GTE // Comparisons don't chain
%left PLUS MINUS
%left MUL DIV MOD
*/
%right UNARY_OP UMINUS // Unary operators (UMINUS for precedence)
%right SAMPLE

%%
// --- Grammar Rules (with position info derived from $N) ---

File:
    DeclarationList {
      ni := NodeInfo{}
      if len($1) > 0 {
        ni.StartPos = $1[0].Pos()
        ni.StopPos = $1[len($1)-1].End()
      }
      SDLlex.(*Lexer).parseResult = &FileDecl{NodeInfo: ni, Declarations: $1}
      // $$ = &File{NodeInfo: ni, Declarations: $1}
    } 
    ;

DeclarationList:
    /* empty */         { $$ = []Node{} }
    | DeclarationList SEMICOLON { $$ = $1 }
    | DeclarationList TopLevelDeclaration {
        $$ = append($1, $2)
    }
    | DeclarationList ImportDecl {
        for _, imp := range $2 {
          $1 = append($1, imp)
        }
        $$ = $1
    }
    ;

TopLevelDeclaration:
      ComponentDecl { $$ = $1 }
    | SystemDecl    { $$ = $1 }
    | AggregatorDecl { $$ = $1 }
    | OptionsDecl   { $$ = $1 }
    | EnumDecl      { $$ = $1 }
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
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $4.(Node).End()), // Pos of OPTIONS, End of RBRACE
            Body: &BlockStmt{
                 NodeInfo: NewNodeInfo($2.(Node).Pos(), $4.(Node).End()),
                 Statements: $3,
             },
         }
    }
    ;

ComponentDecl:
    NATIVE COMPONENT IDENTIFIER LBRACE NativeComponentBodyItemOptList RBRACE { // COMPONENT($1) ... RBRACE($5)
        $$ = &ComponentDecl{
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $6.(Node).End()),
            Name: $3,
            Body: $5,
            IsNative: true,
         }
    }
    | COMPONENT IDENTIFIER LBRACE ComponentBodyItemOptList RBRACE { // COMPONENT($1) ... RBRACE($5)
        $$ = &ComponentDecl{
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $5.(Node).End()),
            Name: $2,
            Body: $4,
         }
    }
    ;

EnumDecl:
    ENUM IDENTIFIER LBRACE CommaIdentifierList RBRACE { // ENUM($1) IDENTIFIER($2) ... RBRACE($5)
        $$ = &EnumDecl{
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $5.(Node).End()),
            Name: $2, // $2 is an IdentifierExpr from lexer, has Pos/End
            Values: $4,
        }
    }
    ;

CommaIdentifierList:
    IDENTIFIER                { $$ = []*IdentifierExpr{$1} }
    | CommaIdentifierList COMMA IDENTIFIER { $$ = append($1, $3) }
    ;

ImportDecl:
    IMPORT ImportList FROM STRING_LITERAL { // IMPORT($1) STRING_LITERAL($2)
        path := $4.(*LiteralExpr)
        for _, imp := range $2 {
          imp.Path = path
        }
        $$ = $2
    }
    ;

ImportList : ImportItem             { $$ = []*ImportDecl{$1}; }
           | ImportList COMMA ImportItem  { $$ = append($$, $3) }
           ;

ImportItem: IDENTIFIER { $$ = &ImportDecl{ImportedItem: $1, Alias: $1 } }
          | IDENTIFIER AS IDENTIFIER { $$ = &ImportDecl{ImportedItem: $1, Alias: $3 } }
          ;

MethodSigDecl:
    IDENTIFIER LPAREN MethodParamListOpt RPAREN { // METHOD($1) ... BlockStmt($6)
        $$ = &MethodDecl{
            NodeInfo: NewNodeInfo($1.Pos(), $4.End()),
            Name: $1,
            Parameters: $3,
        }
    }
    | IDENTIFIER LPAREN MethodParamListOpt RPAREN TypeDecl { // METHOD($1) ... BlockStmt($8)
        $$ = &MethodDecl{
            NodeInfo: NewNodeInfo($1.Pos(), $5.End()),
            Name: $1,
            Parameters: $3,
            ReturnType: $5,
         }
    }
    ;

NativeComponentBodyItemOptList:
                /* empty */ { $$ = []ComponentDeclBodyItem{} }
              | NativeComponentBodyItemList { $$ = $1 }
              ;

NativeComponentBodyItemList:
                NativeComponentBodyItem { $$=[]ComponentDeclBodyItem{$1} }
              | NativeComponentBodyItemList NativeComponentBodyItem { $$=append($1, $2) }
              ;

NativeComponentBodyItem:
      ParamDecl   { $$ = $1 }
    | METHOD MethodSigDecl   { $$ = $2 }
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
    PARAM IDENTIFIER TypeDecl { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()),
            Name: $2,
            TypeDecl: $3, // TypeDecl also needs to have NodeInfo
        }
    }
    | PARAM IDENTIFIER ASSIGN Expression { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $4.End()),
            Name: $2,
            DefaultValue: $4,
        }
    }
    | PARAM IDENTIFIER TypeDecl ASSIGN Expression { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $5.End()),
            Name: $2,
            TypeDecl: $3,
            DefaultValue: $5,
        }
    }
    ;

TypeDecl:
      // PrimitiveType { $$ = $1 } // PrimitiveType actions set NodeInfo
    IDENTIFIER    {
      identNode := $1
      $$ = &TypeDecl{
        NodeInfo: identNode.NodeInfo,
        Name: identNode.Value,
      }
    }
    | LPAREN TypeDeclList RPAREN {  // Tuple type
      if len($2) == 1 {
        $$ = $2[0]
      } else {
        $$ = &TypeDecl{
          NodeInfo: NewNodeInfo($1.Pos(), $3.Pos()),
          Name: "Tuple",
          Args: $2,
        }
      }
    }
    | IDENTIFIER LSQUARE TypeDeclList RSQUARE {
      identNode := $1
      $$ = &TypeDecl{
        NodeInfo: identNode.NodeInfo,
        Name: identNode.Value,
        Args: $3,
      }
    }
    // | QualifiedIdentifier { $$ = $1 } // For future pkg.Type
    // | OutcomeType { $$ = $1 } // Need separate rule if we allow Outcome[T] syntax
    ;

// Placeholder for future pkg.Type or Outcome[T]
// QualifiedIdentifier: IDENTIFIER DOT IDENTIFIER { ... }
// OutcomeType: "Outcome" LBRACKET PrimitiveType RBRACKET { ... }
TypeDeclList:
      TypeDecl { $$ = []*TypeDecl{$1} }
    | TypeDeclList COMMA TypeDecl { $$ = append($1, $3) }
    ;

UsesDecl:
    USES IDENTIFIER IDENTIFIER { // USES($1) ... 
        $$ = &UsesDecl{
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()),
            Name: $2,
            ComponentName: $3,
         }
    }
    // Optional: Add syntax for overrides within uses? Like `uses x: T { p1 = v1; }`
    | USES IDENTIFIER IDENTIFIER LPAREN AssignListOpt RPAREN {
        $$ = &UsesDecl{
             NodeInfo: NewNodeInfo($1.(Node).Pos(), $6.End()),
             Name: $2,
             ComponentName: $3,
             Overrides: $5,
         }
    }
    ;

MethodDecl:
    METHOD MethodSigDecl BlockStmt { // METHOD($1) ... BlockStmt($6)
        $2.Body = $3
        $2.NodeInfo.StopPos = $3.End()
        $$ = $2
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
    IDENTIFIER TypeDecl { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: NewNodeInfo($1.Pos(), $2.End()),
            Name: $1,
            TypeDecl: $2, // TypeDecl also needs to have NodeInfo
        }
    }
    | IDENTIFIER TypeDecl ASSIGN Expression { // PARAM($1) ... 
        $$ = &ParamDecl{
            NodeInfo: NewNodeInfo($1.Pos(), $4.End()),
            Name: $1,
            TypeDecl: $2,
            DefaultValue: $4,
        }
    }
    ;

// --- System ---
SystemDecl:
    SYSTEM IDENTIFIER LBRACE SystemBodyItemOptList RBRACE { // SYSTEM($1) ... RBRACE($5)
        $$ = &SystemDecl{
             NodeInfo: NewNodeInfo($1.(Node).Pos(), $5.(Node).End()),
             Name: $2,
             Body: $4,
        }
    }
    ;

AggregatorDecl:
    NATIVE AGGREGATOR MethodSigDecl { // SYSTEM($1) ... RBRACE($5)
        $$ = &AggregatorDecl{
             NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()),
             Name: $3.Name,
             Parameters: $3.Parameters,
             ReturnType: $3.ReturnType,
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
             NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()),
             Name: $2,
             ComponentName: $3,
             Overrides: []*AssignmentStmt{},
         }
    }
    | USE IDENTIFIER IDENTIFIER LPAREN AssignListOpt RPAREN { // IDENTIFIER($1) ... 
        $$ = &InstanceDecl{
             NodeInfo: NewNodeInfo($1.(Node).Pos(), $6.End()),
             Name: $2,
             ComponentName: $3,
             Overrides: $5,
         }
    }
    ;

AssignListOpt:
      /* empty */  { $$ = []*AssignmentStmt{} }
    | AssignList { $$ = $1 }
    ;

AssignList:
    Assignment             { $$ = []*AssignmentStmt{$1} }
    | AssignList COMMA Assignment { $$ = append($1, $3) }
    ;

Assignment:
    IDENTIFIER ASSIGN Expression { // IDENTIFIER($1) ... 
        $$ = &AssignmentStmt{
            NodeInfo: NewNodeInfo($1.Pos(), $3.End()),
            Var: $1,
            Value: $3,
         }
    }
    ;

/*
AnalyzeDecl:
    ANALYZE IDENTIFIER ASSIGN Expression ExpectBlockOpt { // ANALYZE($1) ... 
        callExpr, ok := $4.(*CallExpr)
        if !ok {
            yyerror(SDLlex, fmt.Sprintf("analyze target must be a method call, found %T at pos %d", $4, $4.(Node).Pos()))
        }
        endPos := 0
        if $5 != nil {
          endPos = $5.End()
        } else {
          endPos = $4.End()
        }
        $$ = &AnalyzeDecl{
             NodeInfo: NewNodeInfo($1.(Node).Pos(), endPos),
             Name: $2,
             Target: callExpr,
             Expectations: $5,
         }
    }
    ;

ExpectBlockOpt:
              { $$ = nil }    // empty
            | ExpectBlock { $$ = $1 };

ExpectBlock:
    EXPECT LBRACE ExpectStmtOptList RBRACE { // EXPECT($1) ... RBRACE($4)
        log.Println("Did Expect Block Hit?")
        $$ = &ExpectationsDecl{
            NodeInfo: NewNodeInfo($1.(Node).Pos(), $4.(Node).End()),
            Expects: $3,
        }
    }
    ;

ExpectStmtOptList:
    { $$ = []*ExpectStmt{} } // empty
    | ExpectStmtList { $$ = $1 }
    ;

ExpectStmtList:
    Expression {
      log.Println("Did we come here????")
      $$ = []*ExpectStmt{$1}
    }
    | ExpectStmtList SEMICOLON Expression {
      log.Println("Why not here Did we come here????")
        cmpExp := $3.(*BinaryExpr);
        expct := &ExpectStmt{ NodeInfo: NewNodeInfo($1[0].Pos(), $3.End()), Target: cmpExp.Left.(*MemberAccessExpr), Operator: cmpExp.Operator, Threshold: cmpExp.Right}
        $$ = append($1, expct)
    }
    ;

ExpectStmt:
    Expression EQ Expression { $$ = &ExpectStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: "==", Threshold: $3} }
    | Expression NEQ Expression { $$ = &ExpectStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: "!=", Threshold: $3} }
    | Expression LT Expression { $$ = &ExpectStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: "<", Threshold: $3} }
    | Expression LTE Expression { $$ = &ExpectStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: "<=", Threshold: $3} }
    | Expression GT Expression { $$ = &ExpectStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: ">", Threshold: $3} }
    | Expression GTE Expression { $$ = &ExpectStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()), Target: $1.(*MemberAccessExpr), Operator: ">=", Threshold: $3} }
    ;
*/

// --- Statements ---
StmtList: 
            /* empty */       { $$ = []Stmt{} }
        | StmtList Stmt {
            $$ = $1
            if $2 != nil {
              $$ = append($$, $2)
            }
        } ;

Stmt:
      LetStmt        { $$ = $1 }
    | ExprStmt       { $$ = $1 }
    | ForStmt       { $$ = $1 }
    | ReturnStmt     { $$ = $1 }
    | IfStmt         { $$ = $1 }
    | DelayStmt      { $$ = $1 }
    | SwitchStmt      { $$ = $1 }
    | BlockStmt      { $$ = $1 }
    | SEMICOLON     { $$ = nil }
    // | AssignStmt     { $$ = $1 } // Disallow assignments as statements? Let's require `let` for now.
    ;

BlockStmt:
    LBRACE StmtList RBRACE {
      $$ = &BlockStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.(Node).End()), Statements: $2 }
    }
    ;

ForStmt: FOR Expression Stmt {
        $$ = &ForStmt{NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()), Condition: $2, Body: $3 }
       }
       ;

LetStmt:
    LET CommaIdentifierList ASSIGN Expression { // LET($1) ... 
         $$ = &LetStmt{
             NodeInfo: NewNodeInfo($1.(Node).Pos(), $4.End()),
             Variables: $2,
             Value: $4,
          }
    }
    ;

/*
AssignStmt: // Rule for simple assignment `a = b;` if needed as statement
    IDENTIFIER ASSIGN Expression {
         // This might conflict with InstanceDecl's Assignment rule if not careful.
         // Let's prefer LetStmt for variables. This rule might be removed.
         // For now, map it to AssignmentStmt AST node used by InstanceDecl.
         $$ = &AssignmentStmt{
             NodeInfo: NewNodeInfo($1.Pos(), $3.Pos()),
             Var: $1,
             Value: $3,
         }
    }
    ;
*/

ExprStmt:
    CallExpr { $$ = &ExprStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $1.(Node).End()), Expression: $1 } }
    | WaitExpr { $$ = &ExprStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $1.(Node).End()), Expression: $1 } }
    ;

ReturnStmt:
    RETURN Expression { $$ = &ReturnStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $2.(Node).End()), ReturnValue: $2 } }
    | RETURN SEMICOLON          { $$ = &ReturnStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $2.(Node).End()), ReturnValue: nil } }
    ;

DelayStmt:
    DELAY Expression { $$ = &DelayStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $2.End()), Duration: $2 } }
    ;

WaitExpr:
    WAIT CommaIdentifierList { // WAIT($1) IDENTIFIER($2) ... 
         idents := $2
         endNode := idents[len(idents)-1] // End at the last identifier in the list
         $$ = &WaitExpr{  Idents: idents }
         $$.(*WaitExpr).NodeInfo = NewNodeInfo($1.Pos(), endNode.End())
    }
    | WAIT CommaIdentifierList USING CallExpr { // WAIT($1) IDENTIFIER($2) ... 
         idents := $2
         endNode := idents[len(idents)-1] // End at the last identifier in the list
         $$ = &WaitExpr{  Idents: idents, Aggregator: $4.(*CallExpr).Function.(*IdentifierExpr), AggregatorParams: $4.(*CallExpr).Args }
         $$.(*WaitExpr).NodeInfo = NewNodeInfo($1.Pos(), endNode.End())
    }
    ;

CommaSepExprListOpt:
      //
      { $$ = []Expr{} }
    | CommaSepExprList { $$ = $1 }
    ;

CommaSepExprList:
    Expression { $$ = []Expr{$1} }
    | CommaSepExprList COMMA Expression { $$ = append($1, $3) }
    ;

// --- Control Flow ---
IfStmt:
    IF Expression BlockStmt IfStmtElseOpt { // IF($1) ...
        endNode := Stmt($3)
        if $4 != nil { endNode = $4 } // End of Else block/IfStmt
        $$ = &IfStmt{
          NodeInfo: NewNodeInfo($1.(Node).Pos(), endNode.End()),
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

SampleExpr:
    SAMPLE Expression { // DISTRIBUTE($1) ... RBRACE($6)
        $$ = &SampleExpr{ FromExpr: $2 }
        $$.NodeInfo = NewNodeInfo($1.(Node).Pos(), $2.(Node).End())
    }
    ;

TotalClauseOpt: /* empty */ { $$=nil } | Expression { $$=$1 };

TupleExpr: LPAREN CommaSepExprList COMMA Expression RPAREN {
          $$ = &TupleExpr{Children: append($2, $4)}
} ;

GoExpr:
    GO BlockStmt { // GO($1) ... BlockStmt($4)
        $$ = &GoExpr{  Stmt: $2 }
        $$.(*GoExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $2.End())
    }
    | GO Expression {
        $$ = &GoExpr{  Expr: $2 }
        $$.(*GoExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $2.End())
    }
    | GOBATCH Expression BlockStmt { // GO($1) ... BlockStmt($4)
        $$ = &GoExpr{  LoopExpr: $2, Stmt: $3 }
        $$.(*GoExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $3.End())
    }
    | GOBATCH Expression Expression {
        $$ = &GoExpr{  LoopExpr: $2, Expr: $3 }
        $$.(*GoExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $3.End())
    }
    ;

// --- Expressions ---
// Expression: OpSepExprList        { $$ = $1 } ;

Expression: ChainedExpr {
        $1.Unchain(nil)
        $$ = $1.UnchainedExpr
    }
    | GoExpr          { $$ = $1 }
    | WaitExpr          { $$ = $1 }
    ;
// Expression: NonAssocBinExpr ;

/*
NonAssocBinExpr:
    BinaryExpr  { $$ = $1 }
    | BinaryExpr BINARY_NC_OP BinaryExpr { 
        $$ = &BinaryExpr{ Left: $1, Operator: $2.String(), Right: $3} 
        $$.(*BinaryExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $3.(Node).End())
    }
    ;

BinaryExpr:
    UnaryExpr  { $$ = $1 }
  | BinaryExpr BINARY_OP UnaryExpr {
     $$ = &BinaryExpr{ Left: $1, Operator: $2.(Node).String(), Right: $3}
     $$.(*BinaryExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $3.(Node).End())
  }
  | BinaryExpr MINUS UnaryExpr {
     $$ = &BinaryExpr{ Left: $1, Operator: $2.(Node).String(), Right: $3}
     $$.(*BinaryExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $3.(Node).End())
  }
  // | For each operator that can be BOTH binary and unary add it here explicitly
*/

ChainedExpr:
    UnaryExpr  {
      $$ = &ChainedExpr{Children: []Expr{$1}}
    }
  | ChainedExpr BINARY_OP UnaryExpr {
      $1.Children = append($1.Children, $3);
      $1.Operators = append($1.Operators, $2.String());
      $$ = $1 ;
  }
  | ChainedExpr MINUS UnaryExpr {
      $1.Children = append($1.Children, $3);
      $1.Operators = append($1.Operators, $2.String());
      $$ = $1 ;
  }
  ;

UnaryExpr: PrimaryExpr { $$=$1 }
    // For Unary, $1 is operator token, $2 is operand Expr node
    | UNARY_OP UnaryExpr { 
        $$ = &UnaryExpr{ Operator: $1.String(), Right: $2} 
        $$.(*UnaryExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $2.(Node).End())
    }
  // | For each operator that can be BOTH binary and unary add it here explicitly
    | MINUS UnaryExpr %prec UMINUS { 
        $$ = &UnaryExpr{ Operator: $1.String(), Right: $2} 
        $$.(*UnaryExpr).NodeInfo = NewNodeInfo($1.(Node).Pos(), $2.(Node).End())
    }
    // For each OPERATOR  that can be binary or unary add a rule like the above MINUS and add a U<OPERATOR> as well
    ;

PrimaryExpr:
       LeafExpr { $$ = $1 }
    | CallExpr                { $$ = $1 }
    ;

LeafExpr:
      LiteralExpr         { $$ = $1 }
    | IDENTIFIER          { $$ = $1 } // Already has position info
    | DistributeExpr      { $$ = $1 } // Expression version
    | SampleExpr          { $$ = $1 }
    | TupleExpr           { $$ = $1 }
    | ParenExpr           { $$ = $1 }
    | MemberAccessExpr        { $$ = $1 }
    | IndexExpr        { $$ = $1 }
    ;

ParenExpr: LPAREN Expression RPAREN { $$ = $2 } // Grouping

LiteralExpr:
      INT_LITERAL             { 
          // SDLlex.(*Lexer).lval)
          $$ = $1 
      }
    | FLOAT_LITERAL           { $$ = $1 }
    | STRING_LITERAL          { $$ = $1 }
    | BOOL_LITERAL            { $$ = $1 }
    | DURATION_LITERAL        { $$ = $1 }
    ;

IndexExpr:
    LeafExpr LSQUARE Expression RSQUARE { // Expression "[" Key "]"
        $$ = &IndexExpr{
             Receiver: $1,
             Key: $3,
        }
        $$.(*IndexExpr).NodeInfo = NewNodeInfo($1.Pos(), $4.End())
    }
    ;

MemberAccessExpr:
    IDENTIFIER DOT IDENTIFIER { // PrimaryExpr($1) DOT($2) IDENTIFIER($3)
        $$ = &MemberAccessExpr{
             Receiver: $1,
             Member: $3,
        }
        $$.(*MemberAccessExpr).NodeInfo = NewNodeInfo($1.Pos(), $3.End())
    }
    | MemberAccessExpr DOT IDENTIFIER { // PrimaryExpr($1) DOT($2) IDENTIFIER($3)
        $$ = &MemberAccessExpr{
            Receiver: $1,
            Member: $3,
        }
        $$.(*MemberAccessExpr).NodeInfo = NewNodeInfo($1.Pos(), $3.End())
    }
    ;

CallExpr:
    IDENTIFIER LPAREN CommaSepExprListOpt RPAREN { // PrimaryExpr($1) LPAREN($2) ArgList($3) RPAREN($4)
        endNode := $4.(Node) // End at RPAREN
        if len($3) > 0 {
             exprList := $3
             endNode = exprList[len(exprList)-1].(Node) // End at last arg
        }
        $$ = &CallExpr{
             Function: $1,
             Args: $3,
        }
        $$.(*CallExpr).NodeInfo = NewNodeInfo($1.Pos(), endNode.End())
    }
    | MemberAccessExpr LPAREN CommaSepExprListOpt RPAREN { // PrimaryExpr($1) LPAREN($2) ArgList($3) RPAREN($4)
         endNode := $4.(Node) // End at RPAREN
         if len($3) > 0 {
             exprList := $3
             endNode = exprList[len(exprList)-1].(Node) // End at last arg
         }
         $$ = &CallExpr{
             Function: $1,
             Args: $3,
         }
        $$.(*CallExpr).NodeInfo = NewNodeInfo($1.Pos(), endNode.End())
    }
    ;

DistributeExpr:
    DISTRIBUTE TotalClauseOpt LBRACE CaseExprListOpt DefaultCaseExprOpt RBRACE {
         $$ = &DistributeExpr{TotalProb: $2, Cases: $4, Default: $5} /* TODO: Pos */
    }
    ;

CaseExprListOpt:
      /* empty */                  { $$ = []*CaseExpr{} }
    | CaseExprList { $$ = $1 }
    ;

CaseExprList:
      CaseExpr { $$ = []*CaseExpr{$1} }
    | CaseExprList CaseExpr { $$ = append($1, $2) }
    ;

CaseExpr:
    Expression ARROW Expression { 
      $$ = &CaseExpr{ Condition: $1, Body: $3 } 
    }
    | Expression ARROW Expression COMMA { // allow optional comma
      $$ = &CaseExpr{ Condition: $1, Body: $3 } 
    }
    ;

DefaultCaseExprOpt:
      /* empty */   { $$ = nil }
    | DefaultCaseExpr { $$ = $1 }
    ;

DefaultCaseExpr:
    DEFAULT ARROW Expression { $$ = $3 }
    | DEFAULT ARROW Expression COMMA { $$ = $3 }
    ;

SwitchStmt:
    SWITCH Expression LBRACE CaseStmtListOpt DefaultCaseStmtOpt RBRACE {
         $$ = &SwitchStmt{Expr: $2, Cases: $4, Default: $5} /* TODO: Pos */
    }
    ;

CaseStmtListOpt:
      /* empty */                  { $$ = []*CaseStmt{} }
    | CaseStmtList { $$ = $1 }
    ;

CaseStmtList:
      CaseStmt { $$ = []*CaseStmt{$1} }
    | CaseStmtList CaseStmt { $$ = append($1, $2) }
    ;

CaseStmt:
    Expression ARROW Stmt { $$ = &CaseStmt{ NodeInfo: NewNodeInfo($1.(Node).Pos(), $3.End()), Condition: $1, Body: $3 } }
    ;

DefaultCaseStmtOpt:
      /* empty */   { $$ = nil }
    | DefaultCaseStmt { $$ = $1 }
    ;

DefaultCaseStmt:
    DEFAULT ARROW Stmt { $$ = $3 }
    ;

%% // --- Go Code Section ---

// Interface for the lexer required by the parser.
type LexerInterface interface {
    Lex(lval *SDLSymType) int
    Error(s string)
    Pos() int       // Start byte position of the last token read
    End() int       // End byte position of the last token read
    Text() string   // Text of the last token read
    Position() (line, col int) // Added: Get line/col of last token start
    LastToken() int // Added: Get the token code that was just lexed
}

// Parse takes an input stream and attempts to parse it according to the SDL grammar. 22222
// It returns the root of the Abstract Syntax Tree (*FileDecl) if successful, or an error.
func Parse(input io.Reader) (*Lexer, *FileDecl, error) {
	// Reset global result before parsing
	lexer := NewLexer(input)
	// Set yyDebug = 3 for verbose parser debugging output
	// yyDebug = 3
	resultCode := SDLParse(lexer) // Call the LALR parser generated by goyacc

	if resultCode != 0 {
		// A syntax error occurred. The lexer's Error method should have been called
		// and stored the error message.
		if lexer.lastError != nil {
			return lexer, nil, lexer.lastError
		}
		// Fallback error message if lexer didn't store one
		return lexer, nil, fmt.Errorf("syntax error near byte %d (Line %d, Col %d)", lexer.location.Pos, lexer.location.Line, lexer.location.Col)
	}

	// Parsing succeeded
	if lexer.parseResult == nil {
		// This indicates a potential issue with the grammar's top rule action
		return lexer, nil, fmt.Errorf("parsing finished successfully, but no AST result was produced")
	}

	return lexer, lexer.parseResult, nil
}

// The parser expects the lexer variable to be named yyLex.
// We can satisfy this by creating a global or passing it via SDLParseWithLexer.
// Using SDLParseWithLexer is cleaner.

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
