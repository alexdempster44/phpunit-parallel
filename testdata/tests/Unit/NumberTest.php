<?php

declare(strict_types=1);

namespace Tests\Unit;

use Tests\TestCase;

class NumberTest extends TestCase
{
    public function testNumberFormat(): void
    {
        $this->pretest();
        $this->assertEquals('1,234.56', number_format(1234.56, 2));
        $this->assertEquals('1 234,56', number_format(1234.56, 2, ',', ' '));
    }

    public function testRoundingModes(): void
    {
        $this->pretest();
        $this->assertEquals(2, round(1.5, 0, PHP_ROUND_HALF_UP));
        $this->assertEquals(2, round(2.5, 0, PHP_ROUND_HALF_EVEN));
    }

    public function testIntegerDivision(): void
    {
        $this->pretest();
        $this->assertEquals(3, intdiv(10, 3));
        $this->assertEquals(1, 10 % 3);
    }

    public function testFloatComparison(): void
    {
        $this->pretest();
        $a = 0.1 + 0.2;
        $b = 0.3;
        $epsilon = 0.00001;

        $this->assertTrue(abs($a - $b) < $epsilon);
    }

    public function testRandomRange(): void
    {
        $this->pretest();
        $random = rand(1, 10);
        $this->assertGreaterThanOrEqual(1, $random);
        $this->assertLessThanOrEqual(10, $random);
    }

    public function testHexConversion(): void
    {
        $this->pretest();
        $this->assertEquals('ff', dechex(255));
        $this->assertEquals(255, hexdec('ff'));
    }

    public function testBinaryConversion(): void
    {
        $this->pretest();
        $this->assertEquals('1010', decbin(10));
        $this->assertEquals(10, bindec('1010'));
    }

    public function testOctalConversion(): void
    {
        $this->pretest();
        $this->assertEquals('12', decoct(10));
        $this->assertEquals(10, octdec('12'));
    }

    public function testIsFinite(): void
    {
        $this->pretest();
        $this->assertTrue(is_finite(123.45));
        $this->assertFalse(is_finite(INF));
        $this->assertFalse(is_finite(NAN));
    }

    public function testBaseConversion(): void
    {
        $this->pretest();
        $this->assertEquals('a', base_convert('10', 10, 16));
        $this->assertEquals('10', base_convert('a', 16, 10));
    }
}
