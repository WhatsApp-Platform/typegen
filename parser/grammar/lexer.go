package grammar

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"text/scanner"
	"unicode"
	
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

// Position represents a position in the source code
type Position struct {
	Filename string
	Line     int
	Column   int
}

func (p Position) String() string {
	if p.Filename != "" {
		return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Keywords maps keyword strings to their token types
var Keywords = map[string]int{
	"import":     IMPORT,
	"struct":     STRUCT,
	"enum":       ENUM,
	"type":       TYPE,
	"const":      CONST,
	
	// Primitive types
	"int8":       INT8,
	"int16":      INT16,
	"int32":      INT32,
	"int64":      INT64,
	"int":        INT,
	"bigint":     BIGINT,
	"nat8":       NAT8,
	"nat16":      NAT16,
	"nat32":      NAT32,
	"nat64":      NAT64,
	"nat":        NAT,
	"bignat":     BIGNAT,
	"float32":    FLOAT32,
	"float64":    FLOAT64,
	"decimal":    DECIMAL,
	"string":     STRING,
	"bool":       BOOL,
	"json":       JSON,
	"time":       TIME,
	"date":       DATE,
	"datetime":   DATETIME,
	"timetz":     TIMETZ,
	"datetz":     DATETZ,
	"datetimetz": DATETIMETZ,
}

// Regular expression for CONSTANT_CASE validation
var constantCaseRegex = regexp.MustCompile(`^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$`)

// IsConstantCase checks if a string follows CONSTANT_CASE convention
func IsConstantCase(name string) bool {
	return constantCaseRegex.MatchString(name)
}

// Lexer implements the goyacc lexer interface
type Lexer struct {
	scanner  scanner.Scanner
	filename string
	result   ast.Node
	errors   []string
}

// NewLexer creates a new lexer for goyacc
func NewLexer(input io.Reader, filename string) *Lexer {
	lex := &Lexer{
		filename: filename,
		errors:   make([]string, 0),
	}
	
	lex.scanner.Init(input)
	lex.scanner.Filename = filename
	lex.scanner.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanStrings | scanner.ScanComments
	
	// Configure scanner for TypeGen syntax
	lex.scanner.IsIdentRune = func(ch rune, i int) bool {
		return unicode.IsLetter(ch) || (i > 0 && (unicode.IsDigit(ch) || ch == '_'))
	}
	
	return lex
}

// Lex implements the goyacc lexer interface
func (l *Lexer) Lex(lval *yySymType) int {
	for {
		ch := l.scanner.Scan()
		pos := Position{
			Filename: l.filename,
			Line:     l.scanner.Line,
			Column:   l.scanner.Column,
		}
		
		switch ch {
		case scanner.EOF:
			return 0
		case scanner.Comment:
			// Skip comments - continue to next token
			continue
		case scanner.Ident:
			text := l.scanner.TokenText()
			if tokenType, exists := Keywords[text]; exists {
				return tokenType
			}
			lval.ident = text
			return IDENTIFIER
		case scanner.Int:
			text := l.scanner.TokenText()
			if val, err := strconv.ParseInt(text, 10, 64); err == nil {
				lval.num = val
				return NUMBER_LITERAL
			}
			l.addError(pos, fmt.Sprintf("invalid number: %s", text))
			continue
		case scanner.String:
			text := l.scanner.TokenText()
			// Remove quotes from string literal
			if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
				unquoted, err := strconv.Unquote(text)
				if err == nil {
					lval.str = unquoted
					return STRING_LITERAL
				}
			}
			l.addError(pos, fmt.Sprintf("invalid string: %s", text))
			continue
		case '{':
			return LBRACE
		case '}':
			return RBRACE
		case '(':
			return LPAREN
		case ')':
			return RPAREN
		case '[':
			return LBRACKET
		case ']':
			return RBRACKET
		case ':':
			return COLON
		case ';':
			return SEMICOLON
		case ',':
			return COMMA
		case '=':
			return EQUALS
		case '?':
			return QUESTION
		case '.':
			return DOT
		default:
			text := l.scanner.TokenText()
			l.addError(pos, fmt.Sprintf("unexpected character: %s", text))
			continue
		}
	}
}

// Error implements the goyacc error interface
func (l *Lexer) Error(s string) {
	pos := Position{
		Filename: l.filename,
		Line:     l.scanner.Line,
		Column:   l.scanner.Column,
	}
	l.errors = append(l.errors, pos.String() + ": " + s)
}

// Result returns the parsed AST
func (l *Lexer) Result() ast.Node {
	return l.result
}

// Errors returns any parse errors
func (l *Lexer) Errors() []string {
	return l.errors
}

// addError adds a lexical error
func (l *Lexer) addError(pos Position, message string) {
	l.errors = append(l.errors, fmt.Sprintf("%s: %s", pos.String(), message))
}

// Parse parses the input using goyacc
func Parse(input io.Reader, filename string) (*Lexer, int) {
	lexer := NewLexer(input, filename)
	result := yyParse(lexer)
	return lexer, result
}