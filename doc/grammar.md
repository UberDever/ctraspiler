```

Source:
    (TopLevelDecl ";")* .
    
TopLevelDecl:
    FunctionDecl .

FunctionDecl:
    "fn" IDENTIFIER Signature FunctionBody? .

Signature:
    "(" IdentifierList? ")" Result? .

Result:
    "" . // there shall be type annotation!

FunctionBody:
    Block .

Block:
    "{" (Statement ";")* "}" .

Statement:
    EmptyStmt
    | IfStmt
    | ReturnStmt
    | Block
    | ExpressionStmt
    | Assignment
    | VarDecl
    | ConstDecl .

EmptyStmt: .

IfStmt:
    "if" Expression Block .

ReturnStmt: 
    "return" ExpressionList .
    
ExpressionStmt:
    Expression .

ConstDecl:
    "const" IdentifierList "=" ExpressionList .

VarDecl:
    "var" IdentifierList "=" ExpressionList .

Assignment:
    ExpressionList AssignOp ExpressionList .

AssignOp:
    "=" .

Expression:
    UnaryExpr | Expression BINARY_OP Expression .

UnaryExpr:
    PrimaryExpr | UNARY_OP UnaryExpr .

PrimaryExpr:
    Operand
    | PrimaryExpr Selector .
    | PrimaryExpr Arguments .

Operand:
    Literal
    | IDENTIFIER
    | "(" Expression ")" .

Literal:
    INT_LIT | FLOAT_LIT | STRING_LIT | BOOL_LIT.

Selector:
    "." IDENTIFIER .

Arguments:
    "(" ExpressionList? ")" .

IdentifierList:
    IDENTIFIER ("," IDENTIFIER)* .

ExpressionList:
    Expression ("," Expression)* .
```