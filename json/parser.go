package json

import (
	"fmt"
	"strconv"
)

func newParser(lex lexer) parser {
	p := parser{
		lex: lex,
	}
	p.readToken()
	p.readToken()

	return p
}

type parser struct {
	lex     lexer
	currTok token
	nextTok token
}

func (p *parser) parse() (value, error) {
	switch p.currTok.kind {
	case tokenLBracket:
		return p.parseArray()
	case tokenLBrace:
		return p.parseObject()
	case tokenNum:
		return p.parseNum()
	case tokenString:
		return p.parseString()
	case tokenBool:
		return p.parseBool()
	default:
		return nil, fmt.Errorf("unknown kind of token: %s", p.currTok.kind)
	}
}

func (p *parser) parseArray() (Array, error) {
	p.readToken()

	var arr Array
	if p.doHaveToken(tokenRBracket) {
		p.readToken()
		return arr, nil
	}
	for {
		val, err := p.parse()
		if err != nil {
			return nil, err
		}
		arr = append(arr, val)

		if !p.doHaveToken(tokenComma) {
			break
		}
		p.readToken()
	}
	if !p.doHaveToken(tokenRBracket) {
		return nil, fmt.Errorf("invalid array format: array should end with ']'")
	}

	p.readToken()

	return arr, nil
}

func (p *parser) parseObject() (Object, error) {
	p.readToken()

	var obj Object
	if p.doHaveToken(tokenRBrace) {
		p.readToken()
		return obj, nil
	}
	for {
		prop, err := p.parseProp()
		if err != nil {
			return nil, fmt.Errorf("failed to parse prop: %w", err)
		}
		obj = append(obj, prop)

		if !p.doHaveToken(tokenComma) {
			break
		}
		p.readToken()
	}
	if !p.doHaveToken(tokenRBrace) {
		return nil, fmt.Errorf("invalid object format: object should end with '}'")
	}

	p.readToken()

	return obj, nil
}

func (p *parser) parseProp() (Prop, error) {
	key, err := p.parseString()
	if err != nil {
		return Prop{}, fmt.Errorf("failed to parse key: %w", err)
	}

	if !p.doHaveToken(tokenColon) {
		return Prop{}, fmt.Errorf("invalid prop format: prop should be composed of key and value separated by ':'")
	}
	p.readToken()

	val, err := p.parse()
	if err != nil {
		return Prop{}, fmt.Errorf("failed to parse value: %w", err)
	}

	return Prop{
		key: key,
		val: val,
	}, nil
}

func (p *parser) parseNum() (Num, error) {
	parsed, err := strconv.ParseInt(p.currTok.literal, 10, 32)
	if err != nil {
		return 0, err
	}

	p.readToken()

	return Num(parsed), nil
}

func (p *parser) parseString() (String, error) {
	if !isStringLiteralQuoted(p.currTok.literal) {
		return "", fmt.Errorf("invalid string format: string should be quoted by '\"'")
	}

	unquoted := unquoteStringLiteral(p.currTok.literal)

	p.readToken()

	return String(unquoted), nil
}

func (p *parser) parseBool() (Bool, error) {
	if p.currTok.literal == literalTrue {
		p.readToken()
		return Bool(true), nil
	}
	if p.currTok.literal == literalFalse {
		p.readToken()
		return Bool(false), nil
	}

	return false, fmt.Errorf("unknown literal of bool: %s", p.currTok.literal)
}

func unquoteStringLiteral(s string) string {
	if !isStringLiteralQuoted(s) {
		return ""
	}

	return s[1 : len(s)-1]
}

func isStringLiteralQuoted(s string) bool {
	if len(s) < 2 {
		return false
	}

	return s[0] == '"' && s[len(s)-1] == '"'
}

func (p *parser) readToken() {
	p.currTok = p.nextTok
	p.nextTok = p.lex.readToken()
}

func (p parser) doHaveToken(kind tokenKind) bool {
	return p.currTok.kind == kind
}

type value interface {
	value()
}

type Num int

func (Num) value() {}

type String string

func (String) value() {}

type Bool bool

func (Bool) value() {}

type Array []value

func (Array) value() {}

type Object []Prop

func (Object) value() {}

type Prop struct {
	key String
	val value
}
