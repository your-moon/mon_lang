#include "tree.h"
#include "defs.h"
#include <wchar.h>

struct ASTnode *mkastnode(int op, struct ASTnode *left,
			  struct ASTnode *right, int intvalue) {
  struct ASTnode *n;

  n = (struct ASTnode *) malloc(sizeof(struct ASTnode));
  if (n == NULL) {
    fprintf(stderr, "Unable to malloc in mkastnode()\n");
    exit(1);
  }

  n->op = op;
  n->left = left;
  n->right = right;
  n->intvalue = intvalue;
  return (n);
}


struct ASTnode *mkastleaf(int intvalue) {
  return (mkastnode(0, NULL, NULL, intvalue));
}

// Make a unary AST node: only one child
struct ASTnode *mkastunary(int op, struct ASTnode *left, int intvalue) {
  return (mkastnode(op, left, NULL, intvalue));
}

void pretty_print_ast(struct ASTnode* node) {
    if (!node) {
        printf("NULL");
        return;
    }

    // If it's a leaf node with an integer value
    if (node->left == NULL && node->right == NULL) {
        wprintf(L"leaf: %d; ", node->intvalue);
    } else {
        // Print the left subtree (recursive call)
        wprintf(L" left: (");
        pretty_print_ast(node->left);

        // Print the operator or operator number
        wprintf(L" op: %d; ", node->op);

        // Print the right subtree (recursive call)
        pretty_print_ast(node->right);
        wprintf(L" :right )\n");
    }
}
