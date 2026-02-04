<?php

declare(strict_types=1);

namespace Tests\Unit;

use Tests\TestCase;

class ExampleTest extends TestCase
{
    public function testTrueIsTrue(): void
    {
        $this->pretest();
        $this->assertTrue(true);
    }

    public function testFalseIsFalse(): void
    {
        $this->pretest();
        $this->assertFalse(false);
    }

    public function testArrayHasKey(): void
    {
        $this->pretest();
        $array = ['foo' => 'bar'];
        $this->assertArrayHasKey('foo', $array);
    }
}
