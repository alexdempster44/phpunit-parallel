<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class HttpClientTest extends TestCase
{
    public function testBuildUrl(): void
    {
        $this->pretest();
        $base = 'https://api.example.com';
        $path = '/users';
        $url = rtrim($base, '/') . '/' . ltrim($path, '/');

        $this->assertEquals('https://api.example.com/users', $url);
    }

    public function testBuildQueryString(): void
    {
        $this->pretest();
        $params = ['page' => 1, 'limit' => 10];
        $query = http_build_query($params);

        $this->assertEquals('page=1&limit=10', $query);
    }

    public function testParseUrl(): void
    {
        $this->pretest();
        $url = 'https://user:pass@example.com:8080/path?query=value#fragment';
        $parts = parse_url($url);

        $this->assertEquals('https', $parts['scheme']);
        $this->assertEquals('example.com', $parts['host']);
        $this->assertEquals(8080, $parts['port']);
        $this->assertEquals('/path', $parts['path']);
    }

    public function testHttpHeaders(): void
    {
        $this->pretest();
        $headers = [
            'Content-Type' => 'application/json',
            'Authorization' => 'Bearer token123',
            'Accept' => 'application/json',
        ];

        $formatted = [];
        foreach ($headers as $key => $value) {
            $formatted[] = "$key: $value";
        }

        $this->assertContains('Content-Type: application/json', $formatted);
    }

    public function testJsonRequestBody(): void
    {
        $this->pretest();
        $data = ['name' => 'John', 'email' => 'john@example.com'];
        $body = json_encode($data);

        $this->assertJson($body);
    }

    public function testFormDataBody(): void
    {
        $this->pretest();
        $data = ['username' => 'john', 'password' => 'secret'];
        $body = http_build_query($data);

        $this->assertEquals('username=john&password=secret', $body);
    }

    public function testStatusCodes(): void
    {
        $this->pretest();
        $codes = [
            200 => 'OK',
            201 => 'Created',
            400 => 'Bad Request',
            401 => 'Unauthorized',
            404 => 'Not Found',
            500 => 'Internal Server Error',
        ];

        $this->assertEquals('OK', $codes[200]);
        $this->assertEquals('Not Found', $codes[404]);
    }

    public function testIsSuccessStatus(): void
    {
        $this->pretest();
        $this->assertTrue($this->isSuccess(200));
        $this->assertTrue($this->isSuccess(201));
        $this->assertFalse($this->isSuccess(400));
        $this->assertFalse($this->isSuccess(500));
    }

    public function testRetryLogic(): void
    {
        $this->pretest();
        $maxRetries = 3;
        $attempts = 0;

        while ($attempts < $maxRetries) {
            $attempts++;
            if ($attempts === 2) {
                break;
            }
        }

        $this->assertEquals(2, $attempts);
    }

    public function testTimeout(): void
    {
        $this->pretest();
        $options = [
            'timeout' => 30,
            'connect_timeout' => 10,
        ];

        $this->assertEquals(30, $options['timeout']);
    }

    private function isSuccess(int $status): bool
    {
        return $status >= 200 && $status < 300;
    }
}
