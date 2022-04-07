package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
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
	return t == nil || (t.Type == TokenType_Unknown && t.Val == "")
}

type AstNode struct {
	Type TokenType
	Val  string

	Left  *AstNode
	Right *AstNode
}

func PrintAst(a *AstNode, lvl int) {

	if a == nil {
		return
	}

	for i := 0; i < lvl; i++ {
		fmt.Print("\u2502")
		fmt.Print("  ")
	}

	fmt.Println("├─'" + a.Val + "'")

	PrintAst(a.Left, lvl+1)
	PrintAst(a.Right, lvl+1)
}

func main() {

	fmt.Println("Please input an equation:")

	reader := bufio.NewReader(os.Stdin)
	eqn, err := reader.ReadString('\n')
	if err != nil {
		panic("Error reading input. Err: " + err.Error())
	}
	// eqn := "+ 5 - 3\n"

	tokens, isInvalid := tokenize(eqn)
	if isInvalid {
		return
	}
	fmt.Printf("\nTokens: %+v\n", tokens)

	if !validateTokens(tokens) {
		return
	}

	//Solve
	ast, err := genAST(tokens)
	if err != nil {
		fmt.Printf("Failed to parse equation. Error: %s\n", err.Error())
		return
	}

	println("Original ast:")
	PrintAst(&ast, 0)

	balancedAst := balanceAst(&ast)
	println("\nBalanced ast:")
	PrintAst(balancedAst, 0)

	ans := solveAst(balancedAst)
	fmt.Printf("\nEqn: %s\n", eqn)
	fmt.Println("Answer:", ans)
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

		//If we have two subsequent numbers and one has a built-in operator then add a plus between them
		prevT := getToken(-1, tokens)
		if t.Type == TokenType_Number && prevT.Type == TokenType_Number {

			if t.Val[0] == '+' || t.Val[0] == '-' {
				tokens = append(tokens, Token{Type: TokenType_Operator, Val: "+"})
			} else {
				fmt.Printf("Error: Two numbers ('%v' and '%v') with no operator between them", prevT.Val, t.Val)
				isInvalid = true
				return
			}

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
			addToken(currToken)
			currToken = Token{}
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

func validateTokens(tokens []Token) bool {

	if len(tokens) == 0 {
		return false
	}

	numberCount := 0
	bracketCount := 0
	operatorCount := 0
	for i := 0; i < len(tokens); i++ {

		t := &tokens[i]

		if t.Type == TokenType_Number {

			numberCount++

			_, err := strconv.ParseFloat(t.Val, 64)
			if err != nil {
				fmt.Printf("Invalid number '%s'\n", t.Val)
				return false
			}
		} else if t.Type == TokenType_OpenBracket {
			bracketCount++
		} else if t.Type == TokenType_CloseBracket {
			bracketCount--
		} else if t.Type == TokenType_Operator {
			operatorCount++
		}

		if bracketCount < 0 {
			fmt.Printf("Can not have a closing bracket before an opening bracket\n")
			return false
		}

		if i == 0 {
			continue
		}

		tPrev := &tokens[i-1]
		if tPrev.Type == TokenType_Operator && t.Type == TokenType_Operator {
			fmt.Printf("Two operators one after the other ('%s' and '%s') are not valid\n", tPrev.Val, t.Val)
			return false
		}
	}

	if bracketCount != 0 {
		fmt.Printf("Not all brackets are closed properly\n")
		return false
	}

	if numberCount < 2 {
		fmt.Printf("Not a valid equation\n")
		return false
	}

	if operatorCount == 0 {
		fmt.Printf("Need at least one operator\n")
		return false
	}

	return true
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

func genAST(tokens []Token) (AstNode, error) {

	n := AstNode{}

	addNode := func(i *int, t, prevT, nextT *Token) error {

		nextAst := &AstNode{Type: nextT.Type, Val: nextT.Val}

		//If we have a bracket on one side then recusrively parse it and add its complete AST
		if nextT.Type == TokenType_OpenBracket {
			x, err := genAST(tokens[*i+2:])
			if err != nil {
				return err
			}
			nextAst.Left = &x

			//Skip brackets that were handled by the recursive solver
			*i += 2
			bracketCount := 1
			for bracketCount != 0 {
				t := &tokens[*i]
				if t.Type == TokenType_OpenBracket {
					bracketCount++
				} else if t.Type == TokenType_CloseBracket {
					bracketCount--
				}

				*i++
			}
			*i--
		}

		if n.Type == TokenType_Unknown {
			n.Type = TokenType_Operator
			n.Val = t.Val
			n.Left = &AstNode{Type: prevT.Type, Val: prevT.Val}
			n.Right = nextAst
		} else {

			oldN := n
			n = AstNode{
				Type:  TokenType_Operator,
				Val:   t.Val,
				Left:  &oldN,
				Right: nextAst,
			}
		}

		return nil
	}

	for i := 0; i < len(tokens); i++ {

		t := &tokens[i]

		if t.Type == TokenType_OpenBracket {

			//Gen ast for whats inside the brackets
			nextAst := &AstNode{Type: TokenType_OpenBracket, Val: t.Val}
			x, err := genAST(tokens[i+1:])
			if err != nil {
				return AstNode{}, err
			}

			nextAst.Left = &x

			//If the first thing we see is a bracket i.e. '(...)' then we handle it as '0 + (...)'
			if n.Type == TokenType_Unknown {
				n.Type = TokenType_Operator
				n.Val = "+"
				n.Left = &AstNode{Type: TokenType_Number, Val: "0"}
				n.Right = nextAst
			} else {

				oldN := n
				n = AstNode{
					Type:  TokenType_Operator,
					Val:   t.Val,
					Left:  &oldN,
					Right: nextAst,
				}
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
			i--
			continue
		}

		if t.Type == TokenType_CloseBracket {
			return n, nil
		}

		if t.Type != TokenType_Operator {
			continue
		}

		//Handle two numbers without an operator
		var prevT *Token
		if i > 0 {
			prevT = getToken(i-1, tokens)
		}

		nextT := getToken(i+1, tokens)
		if nextT.IsEmpty() || prevT.IsEmpty() || prevT.Type == TokenType_Operator || nextT.Type == TokenType_Operator {
			return AstNode{}, errors.New("Operators must be placed between numbers and/or brackets")
		}

		err := addNode(&i, t, prevT, nextT)
		if err != nil {
			return AstNode{}, err
		}
	}

	return n, nil
}

func balanceAst(ast *AstNode) *AstNode {

	if ast == nil {
		return nil
	}

	//Rotate right
	for isParentAstHigherPriority(ast, ast.Left) {
		parent := ast
		child := ast.Left

		parent.Left = child.Right
		child.Right = parent

		ast = child
	}

	//Rotate left
	for isParentAstHigherPriority(ast, ast.Right) {
		parent := ast
		child := ast.Right

		parent.Right = child.Left
		child.Left = parent

		ast = child
	}

	ast.Left = balanceAst(ast.Left)
	ast.Right = balanceAst(ast.Right)

	return ast
}

func isParentAstHigherPriority(parent, child *AstNode) bool {

	if child == nil {
		return false
	}

	if parent.Val == "(" {
		return false
	}

	return (parent.Val == "*" || parent.Val == "/") && (child.Val == "+" || child.Val == "-")
}

func solveAst(ast *AstNode) float64 {

	if ast.Type == TokenType_Number {
		v, _ := strconv.ParseFloat(ast.Val, 64)
		return v
	}

	curr := ast
	switch curr.Val {
	case "+":
		return solveAst(curr.Left) + solveAst(curr.Right)
	case "-":
		return solveAst(curr.Left) - solveAst(curr.Right)
	case "*":
		return solveAst(curr.Left) * solveAst(curr.Right)
	case "/":
		num := solveAst(curr.Left)
		deno := solveAst(curr.Right)
		if deno == 0 {
			fmt.Printf("Can not divide by zero in: '%v / 0'\n", num)
			panic("")
		}
		return num / deno
	case "(":
		return solveAst(curr.Left)
	default:
		panic("Invalid AST node. Value: " + curr.Val)
	}
}
