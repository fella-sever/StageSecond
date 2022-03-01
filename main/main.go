package main

func main() {
	fibonacci(2)

}

func fibonacci(n int) int {
	var (
		num1 int = 1
		num2 int = 1
		sum  int
		i    int
	)
	if n == 1 || n == 2 {
		return 1
	}

	for i = 2; i != n; i++ {
		sum = num1 + num2
		num1 = num2
		num2 = sum

	}
	return sum

}
