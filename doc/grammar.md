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
    | ReturnStmt
    | Block
    | ExpressionStmt
    | Assignment
    | ConstDecl .

EmptyStmt: .

ReturnStmt: 
    "return" ExpressionList .
    
ExpressionStmt:
    Expression .

ConstDecl:
    "const" IdentifierList "=" ExpressionList .

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
    INT_LIT | FLOAT_LIT | STRING_LIT .

Selector:
    "." IDENTIFIER .

Arguments:
    "(" ExpressionList? ")" .

IdentifierList:
    IDENTIFIER ("," IDENTIFIER)* .

ExpressionList:
    Expression ("," Expression)* .
```