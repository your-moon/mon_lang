enum
{
    A_ADD,
    A_SUBTRACT,
    A_MULTIPLY,
    A_DIVIDE,
    A_INTLIT
};

struct ASTnode
{
    int op;
    struct ASTnode *left;
    struct ASTnode *right;
    int intvalue;
};