%{
package grammar

import (
	"fmt"
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)
%}

%union {
	node     ast.Node
	program  *ast.ProgramNode
	decl     ast.Declaration
	decls    []ast.Declaration
	import_  *ast.ImportNode
	imports  []*ast.ImportNode
	struct_  *ast.StructNode
	field    *ast.FieldNode
	fields   []*ast.FieldNode
	enum_    *ast.EnumNode
	variant  *ast.EnumVariantNode
	variants []*ast.EnumVariantNode
	typedef  *ast.TypeAliasNode
	const_   *ast.ConstantNode
	constval ast.ConstantValue
	type_    ast.Type
	ident    string
	str      string
	num      int64
}

%token <ident> IDENTIFIER
%token <str>   STRING_LITERAL
%token <num>   NUMBER_LITERAL

%token IMPORT STRUCT ENUM TYPE CONST
%token LBRACE RBRACE LPAREN RPAREN LBRACKET RBRACKET
%token COLON SEMICOLON COMMA EQUALS QUESTION DOT
%token COMMENT

// Primitive types
%token INT8 INT16 INT32 INT64 INT BIGINT
%token NAT8 NAT16 NAT32 NAT64 NAT BIGNAT
%token FLOAT32 FLOAT64 DECIMAL
%token STRING BOOL JSON
%token TIME DATE DATETIME TIMETZ DATETZ DATETIMETZ

%type <program>  program
%type <imports>  import_list
%type <import_>  import_stmt
%type <str>      module_path qualified_name
%type <decls>    declaration_list
%type <decl>     declaration
%type <struct_>  struct_decl
%type <fields>   field_list non_empty_field_list
%type <field>    field
%type <enum_>    enum_decl
%type <variants> variant_list
%type <variant>  variant
%type <typedef>  type_alias
%type <const_>   const_decl
%type <constval> constant_value
%type <type_>    type_expr primitive_type

%start program

%%

program:
    import_list declaration_list {
        $$ = &ast.ProgramNode{
            Imports:      $1,
            Declarations: $2,
        }
        yylex.(*Lexer).result = $$
    }
|   declaration_list {
        $$ = &ast.ProgramNode{
            Imports:      nil,
            Declarations: $1,
        }
        yylex.(*Lexer).result = $$
    }

import_list:
    import_stmt {
        $$ = []*ast.ImportNode{$1}
    }
|   import_list import_stmt {
        $$ = append($1, $2)
    }

import_stmt:
    IMPORT module_path {
        $$ = &ast.ImportNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Path: $2,
        }
    }

module_path:
    IDENTIFIER {
        $$ = $1
    }
|   module_path DOT IDENTIFIER {
        $$ = $1 + "." + $3
    }

declaration_list:
    declaration {
        $$ = []ast.Declaration{$1}
    }
|   declaration_list declaration {
        $$ = append($1, $2)
    }

declaration:
    struct_decl  { $$ = $1 }
|   enum_decl    { $$ = $1 }
|   type_alias   { $$ = $1 }
|   const_decl   { $$ = $1 }

struct_decl:
    STRUCT IDENTIFIER LBRACE field_list RBRACE {
        $$ = &ast.StructNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name:   $2,
            Fields: $4,
        }
    }

field_list:
    /* empty */ {
        $$ = nil
    }
|   non_empty_field_list {
        $$ = $1
    }

non_empty_field_list:
    field {
        $$ = []*ast.FieldNode{$1}
    }
|   non_empty_field_list field {
        $$ = append($1, $2)
    }

field:
    IDENTIFIER COLON type_expr {
        $$ = &ast.FieldNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name:     $1,
            Type:     $3,
            Optional: false,
        }
    }
|   IDENTIFIER COLON QUESTION type_expr {
        $$ = &ast.FieldNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name:     $1,
            Type:     $4,
            Optional: true,
        }
    }

enum_decl:
    ENUM IDENTIFIER LBRACE variant_list RBRACE {
        $$ = &ast.EnumNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name:     $2,
            Variants: $4,
        }
    }

variant_list:
    variant {
        $$ = []*ast.EnumVariantNode{$1}
    }
|   variant_list variant {
        $$ = append($1, $2)
    }

variant:
    IDENTIFIER {
        $$ = &ast.EnumVariantNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name:    $1,
            Payload: nil,
        }
    }
|   IDENTIFIER COLON type_expr {
        $$ = &ast.EnumVariantNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name:    $1,
            Payload: $3,
        }
    }

type_alias:
    TYPE IDENTIFIER EQUALS type_expr {
        $$ = &ast.TypeAliasNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name: $2,
            Type: $4,
        }
    }

const_decl:
    CONST IDENTIFIER EQUALS constant_value {
        if !IsConstantCase($2) {
            yylex.(*Lexer).Error(fmt.Sprintf("constant name '%s' must be in CONSTANT_CASE format", $2))
            return 1
        }
        $$ = &ast.ConstantNode{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name:  $2,
            Value: $4,
        }
    }

constant_value:
    NUMBER_LITERAL {
        $$ = &ast.IntConstant{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Value: $1,
        }
    }
|   STRING_LITERAL {
        $$ = &ast.StringConstant{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Value: $1,
        }
    }

type_expr:
    primitive_type { $$ = $1 }
|   qualified_name {
        $$ = &ast.NamedType{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            Name: $1,
        }
    }
|   LBRACKET RBRACKET type_expr {
        $$ = &ast.ArrayType{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            ElementType: $3,
        }
    }
|   LBRACKET type_expr RBRACKET type_expr {
        $$ = &ast.MapType{
            BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}},
            KeyType: $2, ValueType: $4,
        }
    }

qualified_name:
    IDENTIFIER {
        $$ = $1
    }
|   qualified_name DOT IDENTIFIER {
        $$ = $1 + "." + $3
    }

primitive_type:
    INT8       { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "int8"} }
|   INT16      { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "int16"} }
|   INT32      { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "int32"} }
|   INT64      { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "int64"} }
|   INT        { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "int"} }
|   BIGINT     { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "bigint"} }
|   NAT8       { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "nat8"} }
|   NAT16      { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "nat16"} }
|   NAT32      { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "nat32"} }
|   NAT64      { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "nat64"} }
|   NAT        { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "nat"} }
|   BIGNAT     { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "bignat"} }
|   FLOAT32    { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "float32"} }
|   FLOAT64    { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "float64"} }
|   DECIMAL    { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "decimal"} }
|   STRING     { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "string"} }
|   BOOL       { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "bool"} }
|   JSON       { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "json"} }
|   TIME       { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "time"} }
|   DATE       { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "date"} }
|   DATETIME   { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "datetime"} }
|   TIMETZ     { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "timetz"} }
|   DATETZ     { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "datetz"} }
|   DATETIMETZ { $$ = &ast.PrimitiveType{BaseNode: ast.BaseNode{Position: ast.Position{Filename: yylex.(*Lexer).filename, Line: yylex.(*Lexer).scanner.Line, Column: yylex.(*Lexer).scanner.Column}}, Name: "datetimetz"} }

%%