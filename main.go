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

type AstType int

const (
	AstType_Unknown AstType = iota
	AstType_Number
	AstType_Operator
)

type AstNode struct {
	Type AstType
	Val  string

	Left  *AstNode
	Right *AstNode
}

//TODO: This is garbage
func (a *AstNode) Print() {

	n := a
	for n != nil {

		fmt.Println("\t", n.Val, "\n/\t\t\\")
		if n.Left != nil {
			fmt.Print(n.Left.Val)
		}

		if n.Right != nil {
			fmt.Println("\t\t", n.Right.Val)
		}

		n = n.Left
	}
}

func main() {

	fmt.Println("Please input an equation:")

	// reader := bufio.NewReader(os.Stdin)
	// eqn, err := reader.ReadString('\n')
	// if err != nil {
	// 	panic("Error reading input. Err: " + err.Error())
	// }
	eqn := "1 + 2 * 3 - 4\n"
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

	a := genAST(tokens)
	fmt.Printf("Eqn: %s\n", eqn)

	println("Original ast:")
	a.Print()

	balancedAst := balanceAst(&a)
	println("Balanced ast:")
	balancedAst.Print()

	ans2 := solveAst(balancedAst)
	println("!!!", ans2)
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

	addToAns := func(f float64, prevToken *Token) {

		if prevToken.Type == TokenType_Operator {

			switch prevToken.Val {
			case "+":
				ans += f
			case "-":
				ans -= f
			case "*":
				ans *= f
			case "/":
				ans /= f
			}

		} else {
			ans += f
		}
	}

	for i := 0; i < len(tokens); i++ {

		t := &tokens[i]

		switch t.Type {

		case TokenType_Number:

			fVal, _ := strconv.ParseFloat(t.Val, 64)
			addToAns(fVal, getToken(i-1, tokens))

		case TokenType_Operator:
		case TokenType_OpenBracket:

			bracketAns := solve(tokens[i+1:])
			addToAns(bracketAns, getToken(i-1, tokens))

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

func genAST(tokens []Token) AstNode {

	n := AstNode{}
	for i := 0; i < len(tokens); i++ {

		t := &tokens[i]
		if t.Type != TokenType_Operator {
			continue
		}

		prevT := getToken(i-1, tokens)
		nextT := getToken(i+1, tokens)
		if nextT.IsEmpty() || (prevT.IsEmpty() && nextT.Type != TokenType_OpenBracket) || prevT.Type == TokenType_Operator || nextT.Type == TokenType_Operator {
			fmt.Println("Operators must be next to numbers or a bracket")
			break
		}

		if n.Type == AstType_Unknown {
			n.Type = AstType_Operator
			n.Val = t.Val
			n.Left = &AstNode{Type: AstType_Number, Val: prevT.Val}
			n.Right = &AstNode{Type: AstType_Number, Val: nextT.Val}
		} else {

			oldN := n
			n = AstNode{
				Type:  AstType_Operator,
				Val:   t.Val,
				Left:  &oldN,
				Right: &AstNode{Type: AstType_Number, Val: nextT.Val},
			}
		}

	}

	return n
}

//TODO: We only balance one level
func balanceAst(ast *AstNode) *AstNode {

	curr := ast
	for curr != nil {

		if isParentAstHigherPriority(ast, ast.Left) {
			parent := ast
			child := ast.Left
			// childChild := ast.Left.Left

			parent.Left = child.Right
			child.Right = parent

			ast = child
			curr = ast.Left
		}

		curr = curr.Left
	}

	return ast
}

func isParentAstHigherPriority(parent, child *AstNode) bool {

	if child == nil {
		return false
	}

	return (parent.Val == "*" || parent.Val == "/") && (child.Val == "+" || child.Val == "-")
}

func solveAst(ast *AstNode) float64 {

	if ast.Type == AstType_Number {
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
		return solveAst(curr.Left) / solveAst(curr.Right)
	default:
	}

	panic("Invalid ast. Value: " + curr.Val)
}
