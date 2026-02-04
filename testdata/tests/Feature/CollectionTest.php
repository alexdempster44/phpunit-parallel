<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;
use ArrayIterator;

class CollectionTest extends TestCase
{
    public function testFirst(): void
    {
        $this->pretest();
        $items = [1, 2, 3];
        $this->assertEquals(1, reset($items));
    }

    public function testLast(): void
    {
        $this->pretest();
        $items = [1, 2, 3];
        $this->assertEquals(3, end($items));
    }

    public function testPluck(): void
    {
        $this->pretest();
        $users = [
            ['name' => 'John', 'age' => 30],
            ['name' => 'Jane', 'age' => 25],
        ];

        $names = array_column($users, 'name');
        $this->assertEquals(['John', 'Jane'], $names);
    }

    public function testKeyBy(): void
    {
        $this->pretest();
        $users = [
            ['id' => 1, 'name' => 'John'],
            ['id' => 2, 'name' => 'Jane'],
        ];

        $keyed = array_column($users, null, 'id');
        $this->assertEquals('John', $keyed[1]['name']);
    }

    public function testGroupBy(): void
    {
        $this->pretest();
        $items = [
            ['type' => 'a', 'value' => 1],
            ['type' => 'b', 'value' => 2],
            ['type' => 'a', 'value' => 3],
        ];

        $grouped = [];
        foreach ($items as $item) {
            $grouped[$item['type']][] = $item;
        }

        $this->assertCount(2, $grouped['a']);
        $this->assertCount(1, $grouped['b']);
    }

    public function testFlatten(): void
    {
        $this->pretest();
        $nested = [[1, 2], [3, 4], [5]];
        $flat = array_merge(...$nested);

        $this->assertEquals([1, 2, 3, 4, 5], $flat);
    }

    public function testChunk(): void
    {
        $this->pretest();
        $items = [1, 2, 3, 4, 5];
        $chunks = array_chunk($items, 2);

        $this->assertEquals([[1, 2], [3, 4], [5]], $chunks);
    }

    public function testZip(): void
    {
        $this->pretest();
        $a = [1, 2, 3];
        $b = ['a', 'b', 'c'];

        $zipped = array_map(null, $a, $b);

        $this->assertEquals([1, 'a'], $zipped[0]);
    }

    public function testPartition(): void
    {
        $this->pretest();
        $numbers = [1, 2, 3, 4, 5, 6];

        $evens = array_filter($numbers, fn($n) => $n % 2 === 0);
        $odds = array_filter($numbers, fn($n) => $n % 2 !== 0);

        $this->assertEquals([2, 4, 6], array_values($evens));
        $this->assertEquals([1, 3, 5], array_values($odds));
    }

    public function testTake(): void
    {
        $this->pretest();
        $items = [1, 2, 3, 4, 5];
        $taken = array_slice($items, 0, 3);

        $this->assertEquals([1, 2, 3], $taken);
    }

    public function testSkip(): void
    {
        $this->pretest();
        $items = [1, 2, 3, 4, 5];
        $skipped = array_slice($items, 2);

        $this->assertEquals([3, 4, 5], $skipped);
    }
}
