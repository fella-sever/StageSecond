package StageSecond

/*функция принимает на вход натуральное число больше единицы, которое означает порядковый номер элемента в ряду
фибоначчи. Функция возвращает число по номеру элемента в ряду фибоначчи*/
func Fibonacci(n int) int {
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
