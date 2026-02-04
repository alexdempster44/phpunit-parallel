<?php

declare(strict_types=1);

namespace Tests\Unit;

use Tests\TestCase;
use Closure;

class ClosureTest extends TestCase
{
    public function testBasicClosure(): void
    {
        $this->pretest();
        $add = fn($a, $b) => $a + $b;
        $this->assertEquals(5, $add(2, 3));
    }

    public function testClosureWithUse(): void
    {
        $this->pretest();
        $multiplier = 3;
        $multiply = fn($n) => $n * $multiplier;

        $this->assertEquals(15, $multiply(5));
    }

    public function testClosureByReference(): void
    {
        $this->pretest();
        $count = 0;
        $increment = function () use (&$count) {
            $count++;
        };

        $increment();
        $increment();
        $this->assertEquals(2, $count);
    }

    public function testClosureCall(): void
    {
        $this->pretest();
        $closure = function () {
            return $this->value;
        };

        $obj = new class {
            public int $value = 42;
        };

        $this->assertEquals(42, $closure->call($obj));
    }

    public function testClosureFromCallable(): void
    {
        $this->pretest();
        $closure = Closure::fromCallable('strtoupper');
        $this->assertEquals('HELLO', $closure('hello'));
    }

    public function testHigherOrderFunction(): void
    {
        $this->pretest();
        $numbers = [1, 2, 3, 4, 5];
        $squared = array_map(fn($n) => $n ** 2, $numbers);

        $this->assertEquals([1, 4, 9, 16, 25], $squared);
    }

    public function testClosureReturningClosure(): void
    {
        $this->pretest();
        $makeAdder = fn($x) => fn($y) => $x + $y;
        $add5 = $makeAdder(5);

        $this->assertEquals(8, $add5(3));
    }

    public function testArrayFilterWithClosure(): void
    {
        $this->pretest();
        $numbers = [1, 2, 3, 4, 5, 6];
        $evens = array_filter($numbers, fn($n) => $n % 2 === 0);

        $this->assertEquals([2, 4, 6], array_values($evens));
    }

    public function testArrayReduceWithClosure(): void
    {
        $this->pretest();
        $numbers = [1, 2, 3, 4];
        $product = array_reduce($numbers, fn($carry, $n) => $carry * $n, 1);

        $this->assertEquals(24, $product);
    }

    public function testUsortWithClosure(): void
    {
        $this->pretest();
        $items = [['name' => 'b'], ['name' => 'a'], ['name' => 'c']];
        usort($items, fn($a, $b) => $a['name'] <=> $b['name']);

        $this->assertEquals('a', $items[0]['name']);
    }
}
