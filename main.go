package main

import (
	"fmt"
	"strconv"
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

type ASTNode struct {
	Type TokenType
}

func main() {

	fmt.Println("Please input an equation:")

	// reader := bufio.NewReader(os.Stdin)
	// eqn, err := reader.ReadString('\n')
	// if err != nil {
	// 	panic("Error reading input. Err: " + err.Error())
	// }
	eqn := "2 * (6 + (4 - 6))\n"
	tokens, isInvalid := tokenize(eqn)
	if isInvalid {
		return
	}
	fmt.Printf("\nTokens: %+v\n", tokens)

	if len(tokens) == 0 {
		return
	}

	//Validate numbers, that no two operators are after each other and brackets
	bracketCount := 0
	for i := 1; i < len(tokens); i++ {

		tPrev := &tokens[i-1]
		t := &tokens[i]

		if t.Type == TokenType_Number {
			_, err := strconv.ParseFloat(t.Val, 64)
			if err != nil {
				fmt.Printf("Invalid number '%s'\n", t.Val)
				return
			}
		}

		if tPrev.Type == TokenType_Operator && t.Type == TokenType_Operator {
			fmt.Printf("Two operators one after the other ('%s' and '%s') are not valid\n", tPrev.Val, t.Val)
			return
		}

		if tPrev.Type == TokenType_OpenBracket {
			bracketCount++
		} else if tPrev.Type == TokenType_CloseBracket {
			bracketCount--
		}

		if bracketCount < 0 {
			fmt.Printf("Can not have a closing bracket before an opening bracket\n")
			return
		}
	}

	tLast := &tokens[len(tokens)-1]

	//Validate last number
	if tLast.Type == TokenType_Number {
		_, err := strconv.ParseFloat(tLast.Val, 64)
		if err != nil {
			fmt.Printf("Invalid number '%s'\n", tLast.Val)
			return
		}
	}

	//Consider ending brackets
	if tLast.Type == TokenType_OpenBracket {
		bracketCount++
	} else if tLast.Type == TokenType_CloseBracket {
		bracketCount--
	}

	if bracketCount != 0 {
		fmt.Printf("Not all brackets are closed properly\n")
		return
	}

	ans := solve(tokens)
	fmt.Println("\nAnswer is:", ans)
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

				var prevChar byte = ' '
				if i > 0 {
					prevChar = eqn[i-1]
				}

				if prevChar == '+' || prevChar == '-' {
					currToken.Val = string(prevChar) + currToken.Val
					tokens = deleteToken(len(tokens)-1, tokens)
				}
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

		case '\r':
			fallthrough
		case '\n':
			addToken(currToken)
			currToken = Token{}
		default:
			isInvalid = true
			fmt.Printf("Invalid char: '%s'\n", string(c))
		}
	}

	return tokens, isInvalid
}

func deleteToken(i int, t []Token) []Token {
	return append(t[:i], t[i+1:]...)
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

func solve(tokens []Token) float64 {

	var ans float64 = 0
	for i := 0; i < len(tokens); i++ {

		t := &tokens[i]

		switch t.Type {

		case TokenType_Number:
			fVal, _ := strconv.ParseFloat(t.Val, 64)

			prevToken := getToken(i-1, tokens)
			if prevToken.Type == TokenType_Operator {

				switch prevToken.Val {
				case "+":
					ans += fVal
				case "-":
					ans -= fVal
				case "*":
					ans *= fVal
				case "/":
					ans /= fVal
				}

			} else {
				ans += fVal
			}

		case TokenType_Operator:
		case TokenType_OpenBracket:

			bracketAns := solve(tokens[i+1:])

			//Add bracket answer to existing answer
			prevToken := getToken(i-1, tokens)
			if prevToken.Type == TokenType_Operator {

				switch prevToken.Val {
				case "+":
					ans += bracketAns
				case "-":
					ans -= bracketAns
				case "*":
					ans *= bracketAns
				case "/":
					ans /= bracketAns
				}

			} else {
				ans += bracketAns
			}

			//Skip brackets that were handled by the recursive solver
			i++
			bracketCount := 1
			for bracketCount != 0 {
				t := &tokens[i]
				if t.Type == TokenType_OpenBracket {
					bracketCount++
				} else if t.Type == TokenType_CloseBracket {
					bracketCount--
				}

				i++
			}

		case TokenType_CloseBracket:
			return ans
		}
	}

	return ans
}
