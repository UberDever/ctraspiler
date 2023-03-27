package main

import (
	"ctranspiler/parser"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

type Parser struct {}

type Listener struct {
	parser.BaseGoParserListener
}

func (l* Listener) VisitTerminal(node antlr.TerminalNode) {
}

func (l *Listener) EnterSourceFile(ctx *parser.SourceFileContext) {
}

func (self *Parser) Parse(data []byte) {
	is := antlr.NewInputStream(string(data))
	lexer := parser.NewGoLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewGoParser(stream)
	listener := &Listener{}
	antlr.ParseTreeWalkerDefault.Walk(listener, parser.SourceFile())
	_ = parser
}

// type TokenType int

// const (
// 	TokenIdentifier = TokenType(tokenizer.TokenKeyword)
// 	TokenString     = TokenType(tokenizer.TokenString)
// 	TokenInteger    = TokenType(tokenizer.TokenInteger)
// 	TokenFloat      = TokenType(tokenizer.TokenFloat)

// 	TokenRawString TokenType = iota
// 	TokenKeyword
// 	TokenOperator
// 	TokenComment

// 	TokenEOF
// 	TokenEnd
// )

// type Token struct {
// 	key   TokenType
// 	value string
// 	line  uint
// }

// func EOF() Token {
// 	return Token{key: TokenEOF}
// }

// func (t Token) isEOF() bool {
// 	return t.key == TokenEOF
// }

// type Parser struct {
// 	config Config
// 	stream *tokenizer.Stream
// }

// func NewParser(config Config) Parser {
// 	return Parser{config: config}
// }

// func (p *Parser) TokenizeFile(filename string) error {
// 	file, err := os.Open(filename)

// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	bytes := make([]byte, 0, 4096)
// 	bytesRead, err := file.Read(bytes)

// 	if err != nil || bytesRead == 0 {
// 		return err
// 	}

// 	p.Tokenize(bytes)
// 	return nil
// }

// func (p *Parser) Tokenize(data []byte) {
// 	t := tokenizer.New()
// 	t.AllowKeywordUnderscore().
// 		AllowNumbersInKeyword().
// 		StopOnUndefinedToken()

// 	t.DefineTokens(tokenizer.TokenKey(TokenKeyword), p.config.keywords())
// 	// t.DefineTokens(tokenizer.TokenKey(TokenOperator), p.config.operators())

// 	t.DefineStringToken(tokenizer.TokenKey(TokenComment), "//", "\n")
// 	t.DefineStringToken(tokenizer.TokenKey(TokenString), "\"", "\"")
// 	t.DefineStringToken(tokenizer.TokenKey(TokenRawString), "`", "`")

// 	p.stream = t.ParseBytes(data)
// }

// // TODO: This won't work even in case of correct match of operators
// // I guess I need to write tokenizer myself after all...
// func (p *Parser) oneLineComment() (Token, bool) {
// 	isOperatorSlash := func(t tokenizer.Token) bool {
// 		return t.Key() == tokenizer.TokenKey(TokenOperator) && bytes.Equal([]byte("/"), t.Value())
// 	}

// 	t := p.stream.CurrentToken()
// 	if !t.IsValid() {
// 		return EOF(), false
// 	}

// 	token := Token{}
// 	if isOperatorSlash(*t) {
// 		t = p.stream.NextToken()
// 		if isOperatorSlash(*t) {
// 			// p.stream.GoNextIfNextIs()
// 		}
// 	}
// 	return token, true
// }

// func (p *Parser) NextToken() Token {
// 	// if t, matched := p.oneLineComment(); matched {
// 	// 	return t
// 	// }

// 	t := p.stream.CurrentToken()
// 	defer p.stream.GoNext()

// 	if !t.IsValid() {
// 		return EOF()
// 	}

// 	token := Token{}
// 	token.key = TokenType(t.Key())
// 	token.line = uint(t.Line())
// 	token.value = string(t.Value())

// 	return token
// }
