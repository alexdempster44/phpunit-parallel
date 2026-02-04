<?php

declare(strict_types=1);

namespace Tests\Unit;

use Tests\TestCase;

class StringTest extends TestCase
{
    public function testStringContains(): void
    {
        $this->pretest();
        $this->assertStringContainsString('foo', 'foobar');
    }

    public function testStringStartsWith(): void
    {
        $this->pretest();
        $this->assertStringStartsWith('Hello', 'Hello World');
    }

    public function testStringEndsWith(): void
    {
        $this->pretest();
        $this->assertStringEndsWith('World', 'Hello World');
    }

    public function testStringLength(): void
    {
        $this->pretest();
        $this->assertEquals(5, strlen('Hello'));
    }

    public function testUpperCase(): void
    {
        $this->pretest();
        $this->assertEquals('HELLO', strtoupper('hello'));
    }

    public function testLowerCase(): void
    {
        $this->pretest();
        $this->assertEquals('hello', strtolower('HELLO'));
    }

    public function testTrim(): void
    {
        $this->pretest();
        $this->assertEquals('hello', trim('  hello  '));
        $this->assertEquals('hello  ', ltrim('  hello  '));
        $this->assertEquals('  hello', rtrim('  hello  '));
    }

    public function testReplace(): void
    {
        $this->pretest();
        $this->assertEquals('Hello Universe', str_replace('World', 'Universe', 'Hello World'));
    }

    public function testSubstring(): void
    {
        $this->pretest();
        $this->assertEquals('Hello', substr('Hello World', 0, 5));
        $this->assertEquals('World', substr('Hello World', 6));
    }

    public function testExplodeImplode(): void
    {
        $this->pretest();
        $parts = explode(',', 'a,b,c');
        $this->assertCount(3, $parts);
        $this->assertEquals('a,b,c', implode(',', $parts));
    }

    public function testPadding(): void
    {
        $this->pretest();
        $this->assertEquals('00042', str_pad('42', 5, '0', STR_PAD_LEFT));
        $this->assertEquals('42000', str_pad('42', 5, '0', STR_PAD_RIGHT));
    }

    public function testWordCount(): void
    {
        $this->pretest();
        $this->assertEquals(3, str_word_count('Hello beautiful world'));
    }
}
