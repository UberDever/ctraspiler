Roadmap for transpiler from high-level language to C.

Main guidelines:

1. Strive for simplicity where can.
1. Strive for performance always. 
1. Use huge front analysis (semantic pass)

## Features

- ~~[] Modules~~
- [] Functions
    - [] Value semantics
    - [] Multiple return
    - [] Anonymous
    - [] ...
- [] Variables
    - [] Declaration
    - [] Consts
    - [] Multiple declarations
    - [] Short declaration
- [] Types
    - [] Basic types
    - [] Type conversion
    - [] Type inference
    - [] Any type
- [] Loops
    - [] Short stmt
    - [] For (c style, while style)
    - [] For range (i, i v, k v)
- [] If
    - [] Short stmt
    - ~~[] Constexpr~~
- [] Switch
    - [] Constant cases (literals)
    - [] Non constant cases
- [] Defer
- [] Pointers
    - [] Auto dereference
- [] Structs
    - [] Initialization
    - [] ...
- [] Arrays
    - [] Static
    - [] Dynamic
    - [] Slicing
- [] Slices
- ~~[] Maps~~
- [] Methods
- ~~[] Interfaces~~
- [] Error handling
    - [] Errors as values
    - [] Optional type
    - [] Union type
- ~~[] Compile time~~
    - [] Generics
    - [] Function evaluation
- [] Memory management
    - [] GC
    - ~~[] Resource heuristics~~
## Scanner

- [x] utf8 chars
- [] Tokens
    - [x] Keywords
    - [x] Identifiers
    - [x] Operators
    - [x] Literals
        - [x] Integer
        - [x] Float decimal
        - [x] Float scientific
        - [x] Strings
- [x] Strip comments
- [x] Autoinsert semicolons

## Parser

- [x] Get EBNF (use go for now)
- [x] Source file
    - [x] Top level decl
        - [x] Declaration
            - [x] Const decl
            - [] Type decl
            - [x] Var decl
        - [x] Function decl
        - [] Method decl
- [x] Statements
    - [x] Decl
    - [] Labeled stmt
    - [x] Simple stmt
        - [x] Expression stmt
        - ~~[] Send stmt~~
        - [] Incdec stmt
        - [x] Assignment 
        - [x] Shortvar decl
    - ~~[] Go stmt~~
    - [x] Return stmt
    - [x] Break stmt
    - [x] Continue stmt
    - [] Goto stmt
    - [] Fallthrough stmt
    - [x] If stmt
    - [] Switch stmt
    - ~~[] Select stmt~~
    - [] For stmt
    - [] Defer stmt

## Results

Implemented: 
1. Scanner with integration of antlr
1. Parser (classic recursive-descend)
1. AST with two forms (non-typed and typed)
1. Identifier lookup analysis
1. Type inference (HM style i guess)

Learned:
1. Tests and infrastructure for compiler
1. Golang quirks and drawbacks
1. How to work with EBNF and write language grammar
1. How to do analysis on AST

Issues and suggestions:
1. Use different language for compiler next time (either functional or full-blown system)
1. Functional type inference haven't made it in full 
1. Functional style is the best for compiler, because we can puzzle up pieces quite easily to test/implement some new stuff