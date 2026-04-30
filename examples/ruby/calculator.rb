class Calculator
  def add(a, b)
    a + b
  end

  def divide(a, b)
    raise ArgumentError, 'division by zero' if b.zero?

    a / b
  end
end
