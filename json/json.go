package json

import (
	"fmt"
	"io"
)

func parse(src string) (value, error) {
	var val scannableVal
	if _, err := fmt.Sscan(src, &val); err != nil {
		return nil, err
	}

	return val.val, nil
}

type value interface {
	value()
}

type scannableVal struct {
	val value
}

func (v *scannableVal) Scan(state fmt.ScanState, _ rune) error {
	peeked, err := peek(state)
	if err != nil {
		return fmt.Errorf("failed to peek character: %w", err)
	}

	switch {
	case canBeObject(string(peeked)):
		var val object
		if _, err := fmt.Fscan(state, &val); err != nil {
			return fmt.Errorf("failed to parse object: %w", err)
		}

		v.val = val

		return nil
	case canBeArray(string(peeked)):
		var val array
		if _, err := fmt.Fscan(state, &val); err != nil {
			return fmt.Errorf("failed to parse array: %w", err)
		}

		v.val = val

		return nil
	case canBeString(string(peeked)):
		var val stringVal
		if _, err := fmt.Fscan(state, &val); err != nil {
			return fmt.Errorf("failed to parse string: %w", err)
		}

		v.val = val

		return nil
	default:
		var val intVal
		if _, err := fmt.Fscan(state, &val); err != nil {
			return fmt.Errorf("failed to parse int: %w", err)
		}

		v.val = val

		return nil
	}
}

func canBeObject(src string) bool {
	if len(src) < 1 {
		return false
	}

	return src[0] == '{'
}

func canBeArray(src string) bool {
	if len(src) < 1 {
		return false
	}

	return src[0] == '['
}

func canBeString(src string) bool {
	if len(src) < 1 {
		return false
	}

	return src[0] == '"'
}

func parseObject(src string) (object, error) {
	var val object
	if _, err := fmt.Sscan(src, &val); err != nil {
		return object{}, err
	}

	return val, nil
}

type object struct {
	props []prop
}

func (object) value() {}

func (o *object) Scan(state fmt.ScanState, _ rune) error {
	peeked, err := peek(state)
	if err != nil {
		return err
	}
	if peeked != '{' {
		return fmt.Errorf("invalid format: object should start with '{'")
	}

	state.ReadRune()

	for {
		peeked, err := peek(state)
		if err != nil {
			return err
		}
		if peeked == '}' {
			break
		}
		state.ReadRune()

		if _, err := state.Token(true, func(r rune) bool {
			return r != '"'
		}); err != nil {
			return err
		}

		var p prop
		if _, err := fmt.Fscan(state, &p); err != nil {
			return fmt.Errorf("failed to parse property: %w", err)
		}

		o.props = append(o.props, p)

		peeked, err = peek(state)
		if err != nil {
			return err
		}
		if peeked != ',' {
			break
		}

		state.ReadRune()
	}

	if _, err := state.Token(true, func(r rune) bool {
		return r != '}'
	}); err != nil {
		return err
	}

	state.ReadRune()

	return nil
}

func parseProp(src string) (prop, error) {
	var p prop
	if _, err := fmt.Sscan(src, &p); err != nil {
		return prop{}, err
	}

	return p, nil
}

type prop struct {
	key string
	val value
}

func (p *prop) Scan(state fmt.ScanState, _ rune) error {
	var key stringVal
	if _, err := fmt.Fscan(state, &key); err != nil {
		return fmt.Errorf("failed to parse key: %w", err)
	}

	peeked, err := peek(state)
	if err != nil {
		return err
	}
	if peeked != ':' {
		return fmt.Errorf("invalid format: prop shuold have ':'")
	}
	state.ReadRune()

	state.SkipSpace()

	var val scannableVal
	if _, err := fmt.Fscan(state, &val); err != nil {
		return fmt.Errorf("failed to parse value: %w", err)
	}

	p.key = string(key)
	p.val = val.val

	return nil
}

func parseArray(src string) (array, error) {
	var val array
	if _, err := fmt.Sscan(src, &val); err != nil {
		return array{}, err
	}

	return val, nil
}

type array []value

func (array) value() {}

func (a *array) Scan(state fmt.ScanState, _ rune) error {
	peeked, err := peek(state)
	if err != nil {
		return err
	}
	if peeked != '[' {
		return fmt.Errorf("invalid format: array should start with '['")
	}

	state.ReadRune()

	for {
		peeked, err := peek(state)
		if err != nil {
			return err
		}
		if peeked == ']' {
			break
		}
		state.ReadRune()

		state.SkipSpace()

		var val scannableVal
		if _, err := fmt.Fscan(state, &val); err != nil {
			return fmt.Errorf("failed to parse value: %w", err)
		}

		*a = append(*a, val.val)

		peeked, err = peek(state)
		if err != nil {
			return err
		}
		if peeked != ',' {
			break
		}
		state.ReadRune()
	}

	if _, err := state.Token(true, func(r rune) bool {
		return r != ']'
	}); err != nil {
		return err
	}

	state.ReadRune()

	return nil
}

func parseInt(src string) (intVal, error) {
	var val intVal
	if _, err := fmt.Sscan(src, &val); err != nil {
		return 0, err
	}

	return val, nil
}

type intVal int

func (intVal) value() {}

func (v *intVal) Scan(state fmt.ScanState, _ rune) error {
	var scanned int
	if _, err := fmt.Fscan(state, &scanned); err != nil {
		return err
	}

	*v = intVal(scanned)

	return nil
}

func parseString(src string) (stringVal, error) {
	var val stringVal
	if _, err := fmt.Sscan(src, &val); err != nil {
		return "", err
	}

	return val, nil
}

type stringVal string

func (stringVal) value() {}

func (v *stringVal) Scan(state fmt.ScanState, _ rune) error {
	peeked, err := peek(state)
	if err != nil {
		return err
	}
	if peeked != '"' {
		return fmt.Errorf("invalid format: string should start with '\"'")
	}

	state.ReadRune()

	read, err := state.Token(false, func(r rune) bool {
		return r != '"'
	})
	if err != nil {
		return err
	}

	state.ReadRune()

	*v = stringVal(read)

	return nil
}

func peek(s io.RuneScanner) (rune, error) {
	r, _, err := s.ReadRune()
	if err != nil {
		return 0, err
	}
	s.UnreadRune()

	return r, nil
}
