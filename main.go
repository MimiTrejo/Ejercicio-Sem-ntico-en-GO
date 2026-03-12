package main

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

type TokenType string

const (
	FOR        TokenType = "FOR"
	IDENT      TokenType = "IDENT"
	ASSIGN     TokenType = "ASSIGN"
	NUMBER     TokenType = "NUMBER"
	LT         TokenType = "LT"
	PLUSPLUS   TokenType = "PLUSPLUS"
	MINUSMINUS TokenType = "MINUSMINUS"
	LBRACE     TokenType = "LBRACE"
	RBRACE     TokenType = "RBRACE"
	SEMI       TokenType = "SEMI"
	ERROR      TokenType = "ERROR"
)

type Token struct {
	Type  TokenType
	Value string
}

func lexer(input string) []Token {

	symbols := []string{";", "{", "}", "<", ":=", "++", "--"}

	for _, s := range symbols {
		input = strings.ReplaceAll(input, s, " "+s+" ")
	}

	parts := strings.Fields(input)

	tokens := []Token{}

	for _, p := range parts {

		switch p {

		case "for":
			tokens = append(tokens, Token{FOR, p})

		case ":=":
			tokens = append(tokens, Token{ASSIGN, p})

		case "<":
			tokens = append(tokens, Token{LT, p})

		case "++":
			tokens = append(tokens, Token{PLUSPLUS, p})

		case "--":
			tokens = append(tokens, Token{MINUSMINUS, p})

		case "{":
			tokens = append(tokens, Token{LBRACE, p})

		case "}":
			tokens = append(tokens, Token{RBRACE, p})

		case ";":
			tokens = append(tokens, Token{SEMI, p})

		default:

			if isNumber(p) {

				tokens = append(tokens, Token{NUMBER, p})

			} else if isIdentifier(p) {

				tokens = append(tokens, Token{IDENT, p})

			} else {

				tokens = append(tokens, Token{ERROR, p})

			}
		}
	}

	return tokens
}

func isNumber(s string) bool {

	if len(s) == 0 {
		return false
	}

	for _, r := range s {

		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func isIdentifier(s string) bool {

	if len(s) == 0 {
		return false
	}

	if !unicode.IsLetter(rune(s[0])) {
		return false
	}

	for _, r := range s {

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func hasLexicalError(tokens []Token) bool {

	for _, t := range tokens {

		if t.Type == ERROR {
			return true
		}
	}

	return false
}

func parseForStmt(tokens []Token) bool {

	pos := 0
	defer func() { recover() }()

	if len(tokens) == 0 || tokens[pos].Type != FOR {
		return false
	}

	pos++

	if !parseInit(tokens, &pos) || tokens[pos].Type != SEMI {
		return false
	}

	pos++

	if !parseCond(tokens, &pos) || tokens[pos].Type != SEMI {
		return false
	}

	pos++

	if !parsePost(tokens, &pos) || !parseBlock(tokens, &pos) {
		return false
	}

	return pos == len(tokens)
}

func parseInit(t []Token, p *int) bool {

	if t[*p].Type == IDENT &&
		t[*p+1].Type == ASSIGN &&
		t[*p+2].Type == NUMBER {

		*p += 3
		return true
	}

	return false
}

func parseCond(t []Token, p *int) bool {

	if t[*p].Type == IDENT &&
		t[*p+1].Type == LT &&
		t[*p+2].Type == NUMBER {

		*p += 3
		return true
	}

	return false
}

func parsePost(t []Token, p *int) bool {

	if t[*p].Type == IDENT &&
		(t[*p+1].Type == PLUSPLUS || t[*p+1].Type == MINUSMINUS) {

		*p += 2
		return true
	}

	return false
}

func parseBlock(t []Token, p *int) bool {

	if t[*p].Type == LBRACE &&
		t[*p+1].Type == RBRACE {

		*p += 2
		return true
	}

	return false
}

func semanticAnalysis(tokens []Token) (bool, string) {

	if len(tokens) < 12 {
		return false, "Código incompleto"
	}

	initVar := tokens[1].Value
	condVar := tokens[5].Value
	postVar := tokens[9].Value

	if initVar != condVar || initVar != postVar {

		return false,
			"Error semántico: la variable del for debe ser la misma en inicialización, condición e incremento/decremento"
	}

	return true, "Semántica correcta"
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		code := r.FormValue("code")
		action := r.FormValue("action")

		var resultHTML string

		if r.Method == "POST" {

			tokens := lexer(code)

			if action == "lexico" {

				resultHTML = "<h3>Tokens:</h3><ul>"

				for _, t := range tokens {

					resultHTML += fmt.Sprintf(
						"<li><b>[%s]</b> : %s</li>",
						t.Type,
						t.Value)
				}

				resultHTML += "</ul>"
			}

			if action == "sintactico" {

				if hasLexicalError(tokens) {

					resultHTML = "<h2 style='color:red'>Error Léxico detectado</h2>"

				} else if parseForStmt(tokens) {

					resultHTML = "<h2 style='color:green'>✔ Sintaxis Correcta</h2>"

				} else {

					resultHTML = "<h2 style='color:red'>✘ Error de Sintaxis</h2>"
				}
			}

			if action == "semantico" {

				if hasLexicalError(tokens) {

					resultHTML = "<h2 style='color:red'>Error Léxico detectado</h2>"

				} else {

					ok, msg := semanticAnalysis(tokens)

					if ok {

						resultHTML = "<h2 style='color:blue'>✔ " + msg + "</h2>"

					} else {

						resultHTML = "<h2 style='color:red'>✘ " + msg + "</h2>"
					}
				}
			}
		}

		fmt.Fprintf(w, `
<html>
<body style="font-family:Arial; margin:40px; max-width:600px;">

<h2>Compilador por Etapas</h2>

<p>Estructura válida:</p>
<code>for i := 0 ; i < 10 ; i ++ { }</code><br>
<code>for i := 10 ; i < 0 ; i -- { }</code>

<form method="POST">

<textarea name="code" rows="3" style="width:100%%">%s</textarea>

<br><br>

<button type="submit" name="action" value="lexico">Analizador Léxico</button>

<button type="submit" name="action" value="sintactico">Analizador Sintáctico</button>

<button type="submit" name="action" value="semantico">Analizador Semántico</button>

</form>

<hr>

%s

</body>
</html>

`, code, resultHTML)

	})

	fmt.Println("Servidor iniciado en http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}
