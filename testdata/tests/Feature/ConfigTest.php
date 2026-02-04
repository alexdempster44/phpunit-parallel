<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class ConfigTest extends TestCase
{
    private array $config = [];

    protected function setUp(): void
    {
        parent::setUp();
        $this->config = [
            'app' => [
                'name' => 'TestApp',
                'debug' => true,
                'timezone' => 'UTC',
            ],
            'database' => [
                'driver' => 'mysql',
                'host' => 'localhost',
                'port' => 3306,
            ],
        ];
    }

    public function testGetValue(): void
    {
        $this->pretest();
        $this->assertEquals('TestApp', $this->get('app.name'));
    }

    public function testGetNestedValue(): void
    {
        $this->pretest();
        $this->assertEquals('localhost', $this->get('database.host'));
    }

    public function testGetWithDefault(): void
    {
        $this->pretest();
        $this->assertEquals('default', $this->get('missing.key', 'default'));
    }

    public function testGetEntireSection(): void
    {
        $this->pretest();
        $app = $this->get('app');

        $this->assertIsArray($app);
        $this->assertArrayHasKey('name', $app);
    }

    public function testSetValue(): void
    {
        $this->pretest();
        $this->set('app.version', '1.0.0');

        $this->assertEquals('1.0.0', $this->get('app.version'));
    }

    public function testHasKey(): void
    {
        $this->pretest();
        $this->assertTrue($this->has('app.name'));
        $this->assertFalse($this->has('app.missing'));
    }

    public function testMergeConfig(): void
    {
        $this->pretest();
        $override = ['app' => ['debug' => false]];
        $merged = array_replace_recursive($this->config, $override);

        $this->assertFalse($merged['app']['debug']);
        $this->assertEquals('TestApp', $merged['app']['name']);
    }

    public function testEnvironmentOverride(): void
    {
        $this->pretest();
        $env = ['APP_DEBUG' => 'false'];
        $config = $this->config;

        if (isset($env['APP_DEBUG'])) {
            $config['app']['debug'] = $env['APP_DEBUG'] === 'true';
        }

        $this->assertFalse($config['app']['debug']);
    }

    public function testAllConfig(): void
    {
        $this->pretest();
        $this->assertArrayHasKey('app', $this->config);
        $this->assertArrayHasKey('database', $this->config);
    }

    public function testConfigKeys(): void
    {
        $this->pretest();
        $keys = array_keys($this->config);
        $this->assertEquals(['app', 'database'], $keys);
    }

    private function get(string $key, mixed $default = null): mixed
    {
        $keys = explode('.', $key);
        $value = $this->config;

        foreach ($keys as $k) {
            if (!isset($value[$k])) {
                return $default;
            }
            $value = $value[$k];
        }

        return $value;
    }

    private function set(string $key, mixed $value): void
    {
        $keys = explode('.', $key);
        $config = &$this->config;

        foreach ($keys as $i => $k) {
            if ($i === count($keys) - 1) {
                $config[$k] = $value;
            } else {
                $config[$k] ??= [];
                $config = &$config[$k];
            }
        }
    }

    private function has(string $key): bool
    {
        return $this->get($key) !== null;
    }
}
