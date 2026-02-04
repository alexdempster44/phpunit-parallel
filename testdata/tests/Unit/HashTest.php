<?php

declare(strict_types=1);

namespace Tests\Unit;

use Tests\TestCase;

class HashTest extends TestCase
{
    public function testMd5(): void
    {
        $this->pretest();
        $hash = md5('hello');
        $this->assertEquals(32, strlen($hash));
        $this->assertEquals('5d41402abc4b2a76b9719d911017c592', $hash);
    }

    public function testSha1(): void
    {
        $this->pretest();
        $hash = sha1('hello');
        $this->assertEquals(40, strlen($hash));
    }

    public function testSha256(): void
    {
        $this->pretest();
        $hash = hash('sha256', 'hello');
        $this->assertEquals(64, strlen($hash));
    }

    public function testSha512(): void
    {
        $this->pretest();
        $hash = hash('sha512', 'hello');
        $this->assertEquals(128, strlen($hash));
    }

    public function testHashHmac(): void
    {
        $this->pretest();
        $hash = hash_hmac('sha256', 'message', 'secret');
        $this->assertEquals(64, strlen($hash));
    }

    public function testHashFile(): void
    {
        $this->pretest();
        $hash = md5_file(__FILE__);
        $this->assertEquals(32, strlen($hash));
    }

    public function testCrc32(): void
    {
        $this->pretest();
        $crc = crc32('hello');
        $this->assertIsInt($crc);
    }

    public function testHashAlgos(): void
    {
        $this->pretest();
        $algos = hash_algos();
        $this->assertContains('sha256', $algos);
        $this->assertContains('md5', $algos);
    }

    public function testHashEquals(): void
    {
        $this->pretest();
        $known = hash('sha256', 'password');
        $user = hash('sha256', 'password');

        $this->assertTrue(hash_equals($known, $user));
    }

    public function testHashPbkdf2(): void
    {
        $this->pretest();
        $hash = hash_pbkdf2('sha256', 'password', 'salt', 1000, 32);
        $this->assertEquals(32, strlen($hash));
    }
}
