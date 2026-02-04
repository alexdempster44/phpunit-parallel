<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;
use SplStack;

class StackTest extends TestCase
{
    public function testPushPop(): void
    {
        $this->pretest();
        $stack = new SplStack();
        $stack->push('first');
        $stack->push('second');

        $this->assertEquals('second', $stack->pop());
        $this->assertEquals('first', $stack->pop());
    }

    public function testLifoOrder(): void
    {
        $this->pretest();
        $stack = new SplStack();
        $stack->push(1);
        $stack->push(2);
        $stack->push(3);

        $result = [];
        while (!$stack->isEmpty()) {
            $result[] = $stack->pop();
        }

        $this->assertEquals([3, 2, 1], $result);
    }

    public function testIsEmpty(): void
    {
        $this->pretest();
        $stack = new SplStack();
        $this->assertTrue($stack->isEmpty());

        $stack->push('item');
        $this->assertFalse($stack->isEmpty());
    }

    public function testCount(): void
    {
        $this->pretest();
        $stack = new SplStack();
        $this->assertEquals(0, $stack->count());

        $stack->push('a');
        $stack->push('b');
        $this->assertEquals(2, $stack->count());
    }

    public function testTop(): void
    {
        $this->pretest();
        $stack = new SplStack();
        $stack->push('first');
        $stack->push('second');

        $this->assertEquals('second', $stack->top());
        $this->assertEquals(2, $stack->count());
    }

    public function testArrayAsStack(): void
    {
        $this->pretest();
        $stack = [];
        $stack[] = 'first';
        $stack[] = 'second';

        $this->assertEquals('second', array_pop($stack));
        $this->assertEquals('first', array_pop($stack));
    }

    public function testUndoStack(): void
    {
        $this->pretest();
        $undoStack = [];
        $undoStack[] = 'action1';
        $undoStack[] = 'action2';
        $undoStack[] = 'action3';

        $lastAction = array_pop($undoStack);
        $this->assertEquals('action3', $lastAction);
    }

    public function testBracketMatching(): void
    {
        $this->pretest();
        $str = '((()))';
        $stack = [];

        foreach (str_split($str) as $char) {
            if ($char === '(') {
                $stack[] = $char;
            } else {
                array_pop($stack);
            }
        }

        $this->assertEmpty($stack);
    }

    public function testReverseWithStack(): void
    {
        $this->pretest();
        $input = [1, 2, 3, 4, 5];
        $stack = [];

        foreach ($input as $item) {
            $stack[] = $item;
        }

        $reversed = [];
        while (!empty($stack)) {
            $reversed[] = array_pop($stack);
        }

        $this->assertEquals([5, 4, 3, 2, 1], $reversed);
    }
}
