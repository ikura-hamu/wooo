package main

import (
	"fmt"
	"slices"
)

type tokenKind int

const (
	// 整数
	TK_INT tokenKind = iota + 1
	// `func`
	TK_FUNC
	// `return`
	TK_RETURN
	// 識別子
	TK_IDENT
	// 区切り記号
	TK_PUNCTUATION
	// 二項演算子
	TK_BINARY_OPERATOR
	// ファイルの終わり
	TK_EOF
)

var keywords = map[tokenKind]string{
	TK_FUNC:   "func",
	TK_RETURN: "return",
}

// token の場所を表す。
// startとposはバイト数 [start, end)
type pos struct {
	fileName string
	start    int
	end      int
}

type token struct {
	kind  tokenKind
	value string
	pos   *pos
}

var letters = []byte{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'_',
}

var digits = []byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
}

var puctuations = []string{
	";", "(", ")", "{", "}",
}

var binaryOperators = []string{
	"+", "-",
}

func isLetterOrDigit(b byte) bool {
	if slices.Contains(letters, b) {
		return true
	}
	return slices.Contains(digits, b)
}

func isKeyword(input string, start int, end int, keyword string) bool {
	if end-start != len(keyword) || end > len(input) {
		return false
	}
	if input[start:end] != keyword {
		return false
	}
	if len(input) > end && isLetterOrDigit(input[end]) {
		return false
	}

	return true
}

func isWhitespaces(input string, start int, end int) bool {
	for i := start; i < end; i++ {
		if input[i] != ' ' && input[i] != '\t' && input[i] != '\n' {
			return false
		}
	}
	return true
}

func tokenizeKeyword(fileName string, input string, start int) (*token, bool) {
	for kind, keyword := range keywords {
		end := start + len(keyword)
		if isKeyword(input, start, end, keyword) {
			return &token{
				kind:  kind,
				value: keyword,
				pos: &pos{
					fileName: fileName,
					start:    start,
					end:      end,
				},
			}, true
		}
	}
	return nil, false
}

func tokenizeIntegerLit(fileName string, input string, start int) (*token, bool, error) {
	if !slices.Contains(digits, input[start]) {
		return nil, false, nil
	}

	// 1ケタだけ
	if input[start] == '0' {
		return &token{
			kind:  TK_INT,
			value: "0",
			pos: &pos{
				fileName: fileName,
				start:    start,
				end:      start + 1,
			},
		}, true, nil
	}

	if input[start+1] != '_' && !slices.Contains(digits, input[start+1]) {
		return &token{
			kind:  TK_INT,
			value: input[start : start+1],
			pos: &pos{
				fileName: fileName,
				start:    start,
				end:      start + 1,
			},
		}, true, nil
	}

	// 2ケタ以上
	decimalDigitsStart := start + 1
	if input[start+1] == '_' {
		decimalDigitsStart = start + 2
	}

	decimalDigitsEnd := decimalDigitsStart + 1
	for decimalDigitsEnd < len(input) {
		if input[decimalDigitsEnd] == '_' {
			decimalDigitsEnd++
			if slices.Contains(digits, input[decimalDigitsEnd]) {
				decimalDigitsEnd++
				continue
			}
			return nil, false, fmt.Errorf("unexpected character in integer literal")
		}

		if slices.Contains(digits, input[decimalDigitsEnd]) {
			decimalDigitsEnd++
			continue
		}

		break
	}

	token := &token{
		kind:  TK_INT,
		value: input[start:decimalDigitsEnd],
		pos: &pos{
			fileName: fileName,
			start:    start,
			end:      decimalDigitsEnd,
		},
	}
	return token, true, nil
}

func tokenizeIdentifier(fileName string, input string, start int) (*token, bool) {
	if !slices.Contains(letters, input[start]) {
		return nil, false
	}
	end := start + 1
	for ; end < len(input) && isLetterOrDigit(input[end]); end++ {
	}
	return &token{
		kind:  TK_IDENT,
		value: input[start:end],
		pos: &pos{
			fileName: fileName,
			start:    start,
			end:      end,
		},
	}, true
}

func tokenizePunctuation(fileName string, input string, start int) (*token, bool) {
	for _, p := range puctuations {
		end := start + len(p)
		if input[start:end] == p {
			return &token{
				kind:  TK_PUNCTUATION,
				value: p,
				pos: &pos{
					fileName: fileName,
					start:    start,
					end:      end,
				},
			}, true
		}
	}
	return nil, false
}

func tokenizeBinaryOperators(fileName string, input string, start int) (*token, bool) {
	for _, op := range binaryOperators {
		end := start + len(op)
		if input[start:end] == op {
			return &token{
				kind:  TK_BINARY_OPERATOR,
				value: op,
				pos: &pos{
					fileName: fileName,
					start:    start,
					end:      end,
				},
			}, true
		}
	}
	return nil, false
}

func isEndOfLine(input string, start int) bool {
	return len(input) <= start || input[start] == '\n'
}

func mustInsertSemicolon(tokens []*token, input string, start int) bool {
	if len(tokens) == 0 {
		return false
	}

	lastToken := tokens[len(tokens)-1]
	if !isEndOfLine(input, start) || lastToken.value == ";" {
		return false
	}

	return lastToken.kind == TK_IDENT ||
		lastToken.kind == TK_INT ||
		lastToken.kind == TK_RETURN ||
		lastToken.kind == TK_PUNCTUATION && lastToken.value == "}"
}

func tokenize(fileName string, input string) ([]*token, error) {
	var tokens []*token
	start := 0
	for end := 1; end <= len(input); {
		if mustInsertSemicolon(tokens, input, start) {
			tokens = append(tokens, &token{
				kind:  TK_PUNCTUATION,
				value: ";",
				pos: &pos{
					fileName: fileName,
					start:    start,
					end:      start,
				},
			})
		}
		if isWhitespaces(input, start, end) {
			start = end
			end = start + 1
			continue
		}
		if keywordToken, ok := tokenizeKeyword(fileName, input, start); ok {
			tokens = append(tokens, keywordToken)
			start = keywordToken.pos.end
			end = start + 1
			continue
		}

		integerToken, ok, err := tokenizeIntegerLit(fileName, input, start)
		if err != nil {
			return nil, err
		}
		if ok {
			tokens = append(tokens, integerToken)
			start = integerToken.pos.end
			end = start + 1
			continue
		}

		if identToken, ok := tokenizeIdentifier(fileName, input, start); ok {
			tokens = append(tokens, identToken)
			start = identToken.pos.end
			end = start + 1
			continue
		}

		if puncToken, ok := tokenizePunctuation(fileName, input, start); ok {
			tokens = append(tokens, puncToken)
			start = puncToken.pos.end
			end = start + 1
			continue
		}

		if binOpToken, ok := tokenizeBinaryOperators(fileName, input, start); ok {
			tokens = append(tokens, binOpToken)
			start = binOpToken.pos.end
			end = start + 1
			continue
		}

		return nil, fmt.Errorf("unexpected character: %c", input[start])
	}

	if mustInsertSemicolon(tokens, input, start) {
		tokens = append(tokens, &token{
			kind:  TK_PUNCTUATION,
			value: ";",
			pos: &pos{
				fileName: fileName,
				start:    start,
				end:      start,
			},
		})
	}

	tokens = append(tokens, &token{
		kind: TK_EOF,
		pos: &pos{
			fileName: fileName,
			start:    start,
			end:      start,
		},
	})

	return tokens, nil
}
