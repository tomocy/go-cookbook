package yaml

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		src      string
		expected value
	}{
		"empty": {
			src:      "",
			expected: Null{},
		},
		"number": {
			src:      "1000",
			expected: Num(1000),
		},
		"string without quotations": {
			src:      "aiueo",
			expected: String(`"aiueo"`),
		},
		"string with quotations": {
			src:      `"aiueo"`,
			expected: String(`"aiueo"`),
		},
		"true": {
			src:      "true",
			expected: Bool(true),
		},
		"false": {
			src:      "false",
			expected: Bool(false),
		},
		"array": {
			src: `- 1
- two
- 3
- "four"
- - one
  - 2
- five: true
- 6: false
  seven: 7`,
			expected: Array{
				Num(1),
				String(`"two"`),
				Num(3),
				String(`"four"`),
				Array{
					String(`"one"`),
					Num(2),
				},
				Dictinary{
					{
						key: String(`"five"`),
						val: Bool(true),
					},
				},
				Dictinary{
					{
						key: String(`"6"`),
						val: Bool(false),
					},
					{
						key: String(`"seven"`),
						val: Num(7),
					},
				},
			},
		},
		"dictionary": {
			src: `a: 1
b: two
c:
   3
e:
	"four"
f:
  g:
    h: i`,
			expected: Dictinary{
				{
					key: String(`"a"`),
					val: Num(1),
				},
				{
					key: String(`"b"`),
					val: String(`"two"`),
				},
				{
					key: String(`"c"`),
					val: Num(3),
				},
				{
					key: String(`"e"`),
					val: String(`"four"`),
				},
				{
					key: String(`"f"`),
					val: Dictinary{
						{
							key: String(`"g"`),
							val: Dictinary{
								{
									key: String(`"h"`),
									val: String(`"i"`),
								},
							},
						},
					},
				},
			},
		},
		"kubernetes": {
			src: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: cookbook
spec:
  replicas: 1
  selector:
    matchLabels:
      app: curl
      version: v1
  template:
    metadata:
      labels:
        app: curl
        version: v1
    spec:
      containers:
      - name: curl
        image: curlimages/curl
        command:
        - /bin/sleep
        - infinity
`,
			expected: Dictinary{
				{
					key: String(`"apiVersion"`),
					val: String(`"apps/v1"`),
				},
				{
					key: String(`"kind"`),
					val: String(`"Deployment"`),
				},
				{
					key: String(`"metadata"`),
					val: Dictinary{
						{
							key: String(`"name"`),
							val: String(`"app"`),
						},
						{
							key: String(`"namespace"`),
							val: String(`"cookbook"`),
						},
					},
				},
				{
					key: String(`"spec"`),
					val: Dictinary{
						{
							key: String(`"replicas"`),
							val: Num(1),
						},
						{
							key: String(`"selector"`),
							val: Dictinary{
								{
									key: String(`"matchLabels"`),
									val: Dictinary{
										{
											key: String(`"app"`),
											val: String(`"curl"`),
										},
										{
											key: String(`"version"`),
											val: String(`"v1"`),
										},
									},
								},
							},
						},
						{
							key: String(`"template"`),
							val: Dictinary{
								{
									key: String(`"metadata"`),
									val: Dictinary{
										{
											key: String(`"labels"`),
											val: Dictinary{
												{
													key: String(`"app"`),
													val: String(`"curl"`),
												},
												{
													key: String(`"version"`),
													val: String(`"v1"`),
												},
											},
										},
									},
								},
								{
									key: String(`"spec"`),
									val: Dictinary{
										{
											key: String(`"containers"`),
											val: Array{
												Dictinary{
													{
														key: String(`"name"`),
														val: String(`"curl"`),
													},
													{
														key: String(`"image"`),
														val: String(`"curlimages/curl"`),
													},
													{
														key: String(`"command"`),
														val: Array{
															String(`"/bin/sleep"`),
															String(`"infinity"`),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			lex := newLexer([]rune(test.src))
			p := newParser(lex)

			actual, err := p.parse()
			if err != nil {
				t.Errorf("should have parse: %s", err)
				return
			}
			if err := assertValue(actual, test.expected); err != nil {
				t.Errorf("unexpected value: %s", err)
				return
			}
		})
	}
}

func assertValue(actual, expected value) error {
	switch expected := expected.(type) {
	case Array:
		if err := assertArray(actual.(Array), expected); err != nil {
			return fmt.Errorf("unexpected array: %w", err)
		}

		return nil
	case Dictinary:
		if err := assertDictinary(actual.(Dictinary), expected); err != nil {
			return fmt.Errorf("unexpected dictionary: %w", err)
		}

		return nil
	default:
		if actual != expected {
			return reprotUnexpected(fmt.Sprintf("%T", expected), actual, expected)
		}

		return nil
	}
}

func assertArray(actual, expected Array) error {
	if len(actual) != len(expected) {
		return reprotUnexpected("len of value", len(actual), len(expected))
	}
	for i, expected := range expected {
		if err := assertValue(actual[i], expected); err != nil {
			return fmt.Errorf("unexpected value: %w", err)
		}
	}

	return nil
}

func assertDictinary(actual, expected Dictinary) error {
	if len(actual) != len(expected) {
		return reprotUnexpected("len of value", len(actual), len(expected))
	}
	for i, expected := range expected {
		if err := assertValue(actual[i].key, expected.key); err != nil {
			return fmt.Errorf("unexpected key: %w", err)
		}
		if err := assertValue(actual[i].val, expected.val); err != nil {
			return fmt.Errorf("unexpected value: %w", err)
		}
	}

	return nil
}
