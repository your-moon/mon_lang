#include "defs.h"
int interpretAST(struct ASTnode *n)
{
    int leftval, rightval;

    if (n->left)
        leftval = interpretAST(n->left);
    if (n->right)
        rightval = interpretAST(n->right);

    switch (n->op)
    {
    case A_ADD:
        return leftval + rightval;
    case A_SUBTRACT:
        return leftval - rightval;
    case A_MULTIPLY:
        return leftval * rightval;
    }
}