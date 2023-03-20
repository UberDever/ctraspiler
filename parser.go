package main

import (
	"os"

	"github.com/bzick/tokenizer"
)

const (
	TokenIdentifier = tokenizer.TokenKeyword
	TokenString     = tokenizer.TokenString
	TokenInteger    = tokenizer.TokenInteger
	TokenFloat      = tokenizer.TokenFloat

	TokenRawString = iota
	TokenKeyword
	TokenOperator
	TokenEOF
	TokenEnd
)

type Token struct {
	key   int
	value string
	line  uint
}

func EOF() Token {
	return Token{key: TokenEOF}
}

func (t Token) isEOF() bool {
	return t.key == TokenEOF
}

type Parser struct {
	config Config
	stream *tokenizer.Stream
}

func NewParser(config Config) Parser {
	return Parser{config: config}
}

func (p *Parser) TokenizeFile(filename string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	bytes := make([]byte, 0, 4096)
	bytesRead, err := file.Read(bytes)

	if err != nil || bytesRead == 0 {
		return err
	}

	p.Tokenize(bytes)
	return nil
}

func (p *Parser) Tokenize(data []byte) {
	t := tokenizer.New()
	t.AllowKeywordUnderscore().
		AllowNumbersInKeyword().
		StopOnUndefinedToken()

	t.DefineStringToken(TokenString, "\"", "\"")
	t.DefineStringToken(TokenRawString, "`", "`")

	t.DefineTokens(TokenKeyword, p.config.keywords())
	t.DefineTokens(TokenOperator, p.config.operators())

	p.stream = t.ParseBytes(data)
}

func (p *Parser) NextToken() Token {
	t := p.stream.CurrentToken()
	defer p.stream.GoNext()

	if !t.IsValid() {
		return EOF()
	}

	token := Token{}
	token.key = int(t.Key())
	token.line = uint(t.Line())
	token.value = string(t.Value())

	return token
}
