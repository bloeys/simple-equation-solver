package main

import (
	"fmt"
	"unicode"
)

type TokenType int

const (
	TokenType_Unknown TokenType = iota
	TokenType_Number
	TokenType_Operator
	TokenType_OpenBracket
	TokenType_CloseBracket
)

type Token struct {
	Val  string
	Type TokenType
}

func (t *Token) IsEmpty() bool {
	return t.Type == TokenType_Unknown && t.Val == ""
}

func main() {

	fmt.Println("Please input an equation:")

	// reader := bufio.NewReader(os.Stdin)
	// eqn, err := reader.ReadString('\n')
	// if err != nil {
	// 	panic("Error reading input. Err: " + err.Error())
	// }
	eqn := "5+3\n"
	tokens, isInvalid := tokenize(eqn)
	if isInvalid {
		return
	}

	fmt.Printf("Tokens: %+v\n", tokens)
}

func tokenize(eqn string) (tokens []Token, isInvalid bool) {

	tokens = make([]Token, 0)

	addToken := func(t Token) {

		if t.IsEmpty() {
			return
		}

		if t.Type == TokenType_Unknown && t.Val != "" {
			fmt.Printf("Invalid character in equation '%v'\n", t.Val)
			isInvalid = true
			return
		}

		tokens = append(tokens, t)
	}

	currToken := Token{}
	for i := 0; i < len(eqn); i++ {

		if isInvalid {
			break
		}

		c := eqn[i]

		//Handle numbers
		if unicode.IsDigit(rune(c)) {

			if currToken.Type == TokenType_Number {
				currToken.Val += string(c)
			} else {
				addToken(currToken)
				currToken = Token{Type: TokenType_Number, Val: string(c)}
			}

			continue
		}

		//Handle others
		switch c {

		case ' ':
			continue

		case '+':
			fallthrough
		case '-':
			fallthrough
		case '*':
			fallthrough
		case '/':
			addToken(currToken)
			addToken(Token{Type: TokenType_Operator, Val: string(c)})
			currToken = Token{}

		case '(':
			addToken(currToken)
			addToken(Token{Type: TokenType_OpenBracket, Val: string(c)})
			currToken = Token{}

		case ')':
			addToken(currToken)
			addToken(Token{Type: TokenType_CloseBracket, Val: string(c)})
			currToken = Token{}
		case '\n':
			addToken(currToken)
			currToken = Token{}
		default:
			fmt.Println("Ignored char:", string(c))
		}
	}

	return tokens, isInvalid
}

func getToken(i int, t []Token) *Token {

	if i >= len(t) || len(t) == 0 {
		return &Token{}
	}

	if i < 0 {
		i = len(t) + i
		if i >= len(t) {
			return &Token{}
		}

		return &t[i]
	}

	return &t[i]
}
