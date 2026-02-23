<?php

declare(strict_types=1);

namespace Tests\Unit;

use Tests\TestCase;

class DeprecationTest extends TestCase
{
    public function testDeprecatedFunction(): void
    {
        trigger_error('This function is deprecated, use newFunction() instead.', E_USER_DEPRECATED);
        $this->assertTrue(true);
    }
}
