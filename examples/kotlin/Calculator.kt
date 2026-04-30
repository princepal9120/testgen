package examples

class Calculator {
    fun add(a: Int, b: Int): Int {
        return a + b
    }

    fun divide(a: Int, b: Int): Int {
        require(b != 0) { "division by zero" }
        return a / b
    }
}
