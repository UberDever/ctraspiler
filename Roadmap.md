
## Scanner

- [] utf8 chars
- [] Tokens
    * [] Keywords
    * [] Identifiers
    * [] Operators
    * [] Literals
        + [] Integer
        + [] Float decimal
        + [] Float scientific
        + [] Strings
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
