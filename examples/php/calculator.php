<?php

function add(int $a, int $b): int
{
    return $a + $b;
}

class Calculator
{
    public function divide(int $a, int $b): int
    {
        if ($b === 0) {
            throw new InvalidArgumentException('division by zero');
        }
        return intdiv($a, $b);
    }
}
