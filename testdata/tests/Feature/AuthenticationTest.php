<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class AuthenticationTest extends TestCase
{
    public function testPasswordHashing(): void
    {
        $this->pretest();
        $password = 'secret123';
        $hash = password_hash($password, PASSWORD_DEFAULT);

        $this->assertTrue(password_verify($password, $hash));
        $this->assertFalse(password_verify('wrongpassword', $hash));
    }

    public function testPasswordHashIsDifferentEachTime(): void
    {
        $this->pretest();
        $password = 'secret123';
        $hash1 = password_hash($password, PASSWORD_DEFAULT);
        $hash2 = password_hash($password, PASSWORD_DEFAULT);

        $this->assertNotEquals($hash1, $hash2);
    }

    public function testTokenGeneration(): void
    {
        $this->pretest();
        $token1 = bin2hex(random_bytes(32));
        $token2 = bin2hex(random_bytes(32));

        $this->assertEquals(64, strlen($token1));
        $this->assertNotEquals($token1, $token2);
    }

    public function testBase64Encoding(): void
    {
        $this->pretest();
        $data = 'username:password';
        $encoded = base64_encode($data);
        $decoded = base64_decode($encoded);

        $this->assertEquals($data, $decoded);
    }

    public function testHashComparison(): void
    {
        $this->pretest();
        $knownHash = hash('sha256', 'secret');
        $userHash = hash('sha256', 'secret');

        $this->assertTrue(hash_equals($knownHash, $userHash));
    }

    public function testSessionIdFormat(): void
    {
        $this->pretest();
        $sessionId = bin2hex(random_bytes(16));
        $this->assertMatchesRegularExpression('/^[a-f0-9]{32}$/', $sessionId);
    }

    public function testJwtStructure(): void
    {
        $this->pretest();
        $header = base64_encode(json_encode(['alg' => 'HS256', 'typ' => 'JWT']));
        $payload = base64_encode(json_encode(['sub' => '1234', 'name' => 'John']));
        $signature = base64_encode('signature');

        $jwt = "$header.$payload.$signature";
        $parts = explode('.', $jwt);

        $this->assertCount(3, $parts);
    }

    public function testApiKeyFormat(): void
    {
        $this->pretest();
        $prefix = 'sk_test_';
        $key = $prefix . bin2hex(random_bytes(24));

        $this->assertStringStartsWith($prefix, $key);
        $this->assertEquals(56, strlen($key));
    }

    public function testCredentialsSanitization(): void
    {
        $this->pretest();
        $username = '  admin  ';
        $sanitized = trim(strtolower($username));

        $this->assertEquals('admin', $sanitized);
    }

    public function testRateLimitingData(): void
    {
        $this->pretest();
        $attempts = [];
        $maxAttempts = 5;
        $userId = 'user123';

        for ($i = 0; $i < 3; $i++) {
            $attempts[$userId] = ($attempts[$userId] ?? 0) + 1;
        }

        $this->assertEquals(3, $attempts[$userId]);
        $this->assertTrue($attempts[$userId] < $maxAttempts);
    }
}
