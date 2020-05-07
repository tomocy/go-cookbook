package yaml

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func parseObject(raw string) (object, error) {
	var obj object
	if _, err := fmt.Sscan(raw, &obj); err != nil {
		return object{}, err
	}

	return obj, nil
}

type object struct {
	fields []field
}

func (o *object) Scan(state fmt.ScanState, _ rune) error {
	for {
		state.SkipSpace()

		isEOF, err := willBeEOF(state)
		if err != nil {
			return err
		}
		if isEOF {
			return nil
		}

		var f field
		if _, err := fmt.Fscan(state, &f); err != nil {
			return err
		}

		o.fields = append(o.fields, f)
	}
}

func willBeEOF(s io.RuneScanner) (bool, error) {
	_, _, err := s.ReadRune()
	if err == nil {
		s.UnreadRune()
		return false, nil
	}
	if err != io.EOF {
		return false, err
	}

	return true, nil
}

type field struct {
	key string
	val value
}

func (f *field) Scan(state fmt.ScanState, _ rune) error {
	var (
		rawKey rawKey
		rawVal rawVal
	)
	if _, err := fmt.Fscan(state, &rawKey, &rawVal); err != nil {
		return err
	}

	f.key = string(rawKey)

	switch {
	case isIntVal(string(rawVal)):
		val, err := parseIntVal(string(rawVal))
		if err != nil {
			return fmt.Errorf("failed to parse int value: %w", err)
		}

		f.val = val

		return nil
	case isStringVal(string(rawVal)):
		val, err := parseStringVal(string(rawVal))
		if err != nil {
			return fmt.Errorf("failed to parse string value: %w", err)
		}

		f.val = val

		return nil
	default:
		var val object
		if _, err := fmt.Fscan(state, &val); err != nil {
			return fmt.Errorf("failed to parse object value: %w", err)
		}

		f.val = val

		return nil
	}
}

type rawKey string

func (k *rawKey) Scan(state fmt.ScanState, _ rune) error {
	tok, err := state.Token(true, func(r rune) bool {
		return r != ':'
	})
	if err != nil {
		return err
	}

	state.ReadRune()

	*k = rawKey(tok)

	return nil
}

type rawVal string

func (v *rawVal) Scan(state fmt.ScanState, verb rune) error {
	tok, err := state.Token(false, func(r rune) bool {
		return r != '\n'
	})
	if err != nil {
		return err
	}

	*v = rawVal(strings.Trim(string(tok), " "))

	return nil
}

func isIntVal(raw string) bool {
	_, err := strconv.ParseInt(raw, 10, 32)
	return err == nil
}

func isStringVal(raw string) bool {
	if len(raw) == 0 {
		return false
	}

	return raw[0] == '"' && raw[len(raw)-1] == '"'
}

func parseStringVal(raw string) (stringVal, error) {
	val := strings.Trim(raw, "\"")
	return stringVal(val), nil
}

func parseIntVal(raw string) (intVal, error) {
	val, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, err
	}

	return intVal(val), nil
}

type value interface {
	value()
}

func (o object) value() {}

type intVal int

func (v intVal) value() {}

type stringVal string

func (v stringVal) value() {}
