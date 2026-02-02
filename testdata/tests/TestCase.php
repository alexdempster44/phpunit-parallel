<?php

declare(strict_types=1);

namespace Tests;

use \Exception;
use PHPUnit\Framework\TestCase as BaseTestCase;

class TestCase extends BaseTestCase
{
    protected function setUp(): void
    {
        if (rand(0, 32) === 0) {
            throw new Exception();
        }

        usleep((int) (1_000_000 * 0.25));
    }
}
