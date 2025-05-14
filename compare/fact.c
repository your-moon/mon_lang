#include <stdio.h>
#include <time.h>

void print_result(long result) {
    printf("%ld\n", result);
}

long factorial(int n) {
    if (n <= 1) {
        return 1;
    }

    long result = n * factorial(n - 1);
    print_result(result);
    return result;
}

int main() {
    // Measure execution time
    clock_t start = clock();

    long result = factorial(12);  // 12! = 479001600

    clock_t end = clock();
    double elapsed_time = (double)(end - start) / CLOCKS_PER_SEC;

    printf("Final result: %ld\n", result);
    printf("Execution time: %.9f seconds\n", elapsed_time);

    return 0;
}
