#include <stdio.h>
#include <stdlib.h>
struct ASTnode *mkastnode(int op, struct ASTnode *left,
			  struct ASTnode *right, int intvalue);
struct ASTnode *mkastleaf(int intvalue);
struct ASTnode *mkastunary(int op, struct ASTnode *left, int intvalue);
void pretty_print_ast(struct ASTnode* node);
