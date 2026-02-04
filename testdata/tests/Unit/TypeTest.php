<?php

declare(strict_types=1);

namespace Tests\Unit;

use Tests\TestCase;

class TypeTest extends TestCase
{
    public function testGetType(): void
    {
        $this->pretest();
        $this->assertEquals('integer', gettype(42));
        $this->assertEquals('double', gettype(3.14));
        $this->assertEquals('string', gettype('hello'));
        $this->assertEquals('array', gettype([]));
        $this->assertEquals('boolean', gettype(true));
        $this->assertEquals('NULL', gettype(null));
    }

    public function testIsArray(): void
    {
        $this->pretest();
        $this->assertTrue(is_array([]));
        $this->assertTrue(is_array([1, 2, 3]));
        $this->assertFalse(is_array('string'));
    }

    public function testIsString(): void
    {
        $this->pretest();
        $this->assertTrue(is_string('hello'));
        $this->assertTrue(is_string(''));
        $this->assertFalse(is_string(123));
    }

    public function testIsInt(): void
    {
        $this->pretest();
        $this->assertTrue(is_int(42));
        $this->assertFalse(is_int(42.0));
        $this->assertFalse(is_int('42'));
    }

    public function testIsFloat(): void
    {
        $this->pretest();
        $this->assertTrue(is_float(3.14));
        $this->assertTrue(is_float(1.0));
        $this->assertFalse(is_float(1));
    }

    public function testIsBool(): void
    {
        $this->pretest();
        $this->assertTrue(is_bool(true));
        $this->assertTrue(is_bool(false));
        $this->assertFalse(is_bool(1));
        $this->assertFalse(is_bool('true'));
    }

    public function testIsNull(): void
    {
        $this->pretest();
        $this->assertTrue(is_null(null));
        $this->assertFalse(is_null(''));
        $this->assertFalse(is_null(0));
    }

    public function testIsCallable(): void
    {
        $this->pretest();
        $this->assertTrue(is_callable('strlen'));
        $this->assertTrue(is_callable(fn() => true));
        $this->assertFalse(is_callable('nonexistent_function'));
    }

    public function testIsIterable(): void
    {
        $this->pretest();
        $this->assertTrue(is_iterable([1, 2, 3]));
        $this->assertTrue(is_iterable(new \ArrayIterator([1, 2, 3])));
        $this->assertFalse(is_iterable('string'));
    }

    public function testTypeCasting(): void
    {
        $this->pretest();
        $this->assertEquals(42, (int) '42');
        $this->assertEquals(3.14, (float) '3.14');
        $this->assertEquals('42', (string) 42);
        $this->assertEquals([42], (array) 42);
        $this->assertTrue((bool) 1);
        $this->assertFalse((bool) 0);
    }
}
