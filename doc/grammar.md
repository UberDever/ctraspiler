```

Source:
    TopLevelDecl* ";" .
    
TopLevelDecl:
    FunctionDecl .

FunctionDecl:
    "func" IDENTIFIER Signature FunctionBody? .

Signature:
    Parameters Result? .

Parameters:
    "(" IdentifierList? ")" .

Result:
    "" . // there shall be type annotation!

FunctionBody:
    Block .

Block:
    "{" Statement* ";" "}" .

Statement:
    EmptyStmt
    | ExpressionStmt
    | Assignment
    | DeclStmt .

EmptyStmt: .
    
ExpressionStmt:
    Expression .

DeclStmt:
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
    | PrimaryExpr Arguments .

Operand:
    Literal
    | IDENTIFIER
    | "(" Expression ")"

Literal:
    INT_LIT | FLOAT_LIT | STRING_LIT .

Arguments:
    "(" ExpressionList? ","? ")"

IdentifierList:
    IDENTIFIER ("," IDENTIFIER)*

ExpressionList:
    Expression ("," Expression)*

```