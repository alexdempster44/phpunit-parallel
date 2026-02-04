<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class LoggingTest extends TestCase
{
    private array $logs = [];

    public function testLogMessage(): void
    {
        $this->pretest();
        $this->log('info', 'Test message');

        $this->assertCount(1, $this->logs);
        $this->assertEquals('Test message', $this->logs[0]['message']);
    }

    public function testLogLevel(): void
    {
        $this->pretest();
        $this->log('error', 'Error occurred');

        $this->assertEquals('error', $this->logs[0]['level']);
    }

    public function testLogContext(): void
    {
        $this->pretest();
        $this->log('info', 'User logged in', ['user_id' => 42]);

        $this->assertEquals(42, $this->logs[0]['context']['user_id']);
    }

    public function testLogTimestamp(): void
    {
        $this->pretest();
        $this->log('debug', 'Debug message');

        $this->assertArrayHasKey('timestamp', $this->logs[0]);
    }

    public function testLogLevels(): void
    {
        $this->pretest();
        $levels = ['emergency', 'alert', 'critical', 'error', 'warning', 'notice', 'info', 'debug'];

        foreach ($levels as $level) {
            $this->log($level, "Message at $level");
        }

        $this->assertCount(8, $this->logs);
    }

    public function testFilterByLevel(): void
    {
        $this->pretest();
        $this->log('info', 'Info 1');
        $this->log('error', 'Error 1');
        $this->log('info', 'Info 2');

        $errors = array_filter($this->logs, fn($l) => $l['level'] === 'error');
        $this->assertCount(1, $errors);
    }

    public function testLogFormatting(): void
    {
        $this->pretest();
        $this->log('info', 'User {name} logged in', ['name' => 'John']);

        $formatted = $this->formatMessage($this->logs[0]);
        $this->assertStringContainsString('John', $formatted);
    }

    public function testLogChannel(): void
    {
        $this->pretest();
        $this->logs[] = [
            'channel' => 'auth',
            'level' => 'info',
            'message' => 'Login attempt',
            'timestamp' => time(),
            'context' => [],
        ];

        $this->assertEquals('auth', $this->logs[0]['channel']);
    }

    public function testClearLogs(): void
    {
        $this->pretest();
        $this->log('info', 'Message');
        $this->logs = [];

        $this->assertEmpty($this->logs);
    }

    public function testLogCount(): void
    {
        $this->pretest();
        for ($i = 0; $i < 5; $i++) {
            $this->log('info', "Message $i");
        }

        $this->assertCount(5, $this->logs);
    }

    private function log(string $level, string $message, array $context = []): void
    {
        $this->logs[] = [
            'level' => $level,
            'message' => $message,
            'context' => $context,
            'timestamp' => time(),
        ];
    }

    private function formatMessage(array $log): string
    {
        $message = $log['message'];
        foreach ($log['context'] as $key => $value) {
            $message = str_replace("{{$key}}", (string) $value, $message);
        }
        return "[{$log['level']}] $message";
    }
}
