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

- [] utf8 chars
- [] Tokens
    - [] Keywords
    - [] Identifiers
    - [] Operators
    - [] Literals
        - [] Integer
        - [] Float decimal
        - [] Float scientific
        - [] Strings
- [] Strip comments

## Parser

- [] Get EBNF (use go for now)
- [] Source file
    - [] Top level decl
        - [] Declaration
            - [] Const decl
            - [] Type decl
            - [] Var decl
        - [] Function decl
        - [] Method decl
- [] Statements
    - [] Decl
    - [] Labeled stmt
    - [] Simple stmt
        - [] Expression stmt
        - ~~[] Send stmt~~
        - [] Incdec stmt
        - [] Assignment 
        - [] Shortvar decl
    - ~~[] Go stmt~~
    - [] Return stmt
    - [] Break stmt
    - [] Continue stmt
    - [] Goto stmt
    - [] Fallthrough stmt
    - [] If stmt
    - [] Switch stmt
    - ~~[] Select stmt~~
    - [] For stmt
    - [] Defer stmt
