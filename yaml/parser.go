package yaml

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
	lex              lexer
	currTok, nextTok token
}

func (p *parser) parse() (value, error) {
	switch p.currTok.kind {
	case tokenEOF:
		return Null{}, nil
	case tokenHyphen:
		val, err := p.parseArray()
		if err != nil {
			return nil, fmt.Errorf("failed to parse array: %w", err)
		}

		return val, nil
	case tokenNum, tokenString, tokenBool:
		if p.willHaveToken(tokenColon) {
			val, err := p.parseObject()
			if err != nil {
				return nil, fmt.Errorf("failed to parse object: %w", err)
			}

			return val, nil
		}

		switch p.currTok.kind {
		case tokenNum:
			val, err := p.parseNum()
			if err != nil {
				return nil, fmt.Errorf("failed to parse number: %w", err)
			}

			return val, nil
		case tokenString:
			val, err := p.parseString()
			if err != nil {
				return nil, fmt.Errorf("failed to parse string: %w", err)
			}

			return val, nil
		case tokenBool:
			val, err := p.parseBool()
			if err != nil {
				return nil, fmt.Errorf("failed to parse bool: %w", err)
			}

			return val, nil
		default:
			return nil, fmt.Errorf("unknown type of token: %s", p.currTok.kind)
		}
	default:
		return nil, fmt.Errorf("unknown type of token: %s", p.currTok.kind)
	}
}

func (p *parser) parseArray() (Array, error) {
	basePos := p.currTok.pos
	p.readToken()

	var arr Array
	for {
		val, err := p.parse()
		if err != nil {
			return nil, fmt.Errorf("failed to parse value: %w", err)
		}

		arr = append(arr, val)

		if !p.doHaveTokenInBase(tokenHyphen, basePos.start) {
			break
		}
		p.readToken()
	}

	return arr, nil
}

func (p *parser) parseObject() (Object, error) {
	basePos := p.currTok.pos

	var obj Object
	for {
		prop, err := p.parseProp()
		if err != nil {
			return nil, fmt.Errorf("failed to parse prop: %w", err)
		}

		obj = append(obj, prop)

		if !p.doHaveTokenInBase(tokenString, basePos.start) || !p.willHaveToken(tokenColon) {
			break
		}
	}

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
	quoted := p.currTok.literal
	if !isStringQuoted(quoted) {
		quoted = fmt.Sprintf("\"%s\"", quoted)
	}

	p.readToken()

	return String(quoted), nil
}

func isStringQuoted(s string) bool {
	if len(s) < 2 {
		return false
	}

	return s[0] == '"' && s[len(s)-1] == '"'
}

func (p *parser) parseBool() (Bool, error) {
	l := p.currTok.literal
	if l == literalTrue {
		p.readToken()
		return true, nil
	}
	if l == literalFalse {
		p.readToken()
		return false, nil
	}

	return false, fmt.Errorf("invalid literal of bool: %s", l)
}

func (p *parser) readToken() {
	p.currTok = p.nextTok
	p.nextTok = p.lex.readToken()
}

func (p parser) doHaveTokenInBase(kind tokenKind, base int) bool {
	return p.doHaveToken(kind) && p.currTok.pos.start == base
}

func (p parser) doHaveToken(kind tokenKind) bool {
	return p.currTok.kind == kind
}

func (p parser) willHaveToken(kind tokenKind) bool {
	return p.nextTok.kind == kind
}

type value interface {
	value()
}

type Null struct{}

func (Null) value() {}

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
