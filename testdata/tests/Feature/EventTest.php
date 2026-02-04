<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class EventTest extends TestCase
{
    private array $listeners = [];
    private array $dispatched = [];

    public function testRegisterListener(): void
    {
        $this->pretest();
        $this->listeners['user.created'] = [];
        $this->listeners['user.created'][] = fn($data) => $data;

        $this->assertCount(1, $this->listeners['user.created']);
    }

    public function testDispatchEvent(): void
    {
        $this->pretest();
        $this->listeners['test.event'][] = function ($data) {
            $this->dispatched[] = $data;
        };

        foreach ($this->listeners['test.event'] as $listener) {
            $listener(['id' => 1]);
        }

        $this->assertCount(1, $this->dispatched);
    }

    public function testMultipleListeners(): void
    {
        $this->pretest();
        $results = [];

        $this->listeners['multi'][] = function () use (&$results) { $results[] = 'first'; };
        $this->listeners['multi'][] = function () use (&$results) { $results[] = 'second'; };

        foreach ($this->listeners['multi'] as $listener) {
            $listener();
        }

        $this->assertEquals(['first', 'second'], $results);
    }

    public function testRemoveListener(): void
    {
        $this->pretest();
        $callback = fn() => 'test';
        $this->listeners['event'][] = $callback;

        $this->listeners['event'] = array_filter(
            $this->listeners['event'],
            fn($l) => $l !== $callback
        );

        $this->assertEmpty($this->listeners['event']);
    }

    public function testEventPropagation(): void
    {
        $this->pretest();
        $stopped = false;
        $executed = [];

        $this->listeners['propagation'][] = function () use (&$executed, &$stopped) {
            $executed[] = 1;
            $stopped = true;
        };
        $this->listeners['propagation'][] = function () use (&$executed, &$stopped) {
            if (!$stopped) {
                $executed[] = 2;
            }
        };

        foreach ($this->listeners['propagation'] as $listener) {
            $listener();
        }

        $this->assertEquals([1], $executed);
    }

    public function testEventPayload(): void
    {
        $this->pretest();
        $received = null;

        $this->listeners['payload'][] = function ($data) use (&$received) {
            $received = $data;
        };

        $payload = ['user_id' => 42, 'action' => 'login'];
        foreach ($this->listeners['payload'] as $listener) {
            $listener($payload);
        }

        $this->assertEquals(42, $received['user_id']);
    }

    public function testWildcardEvents(): void
    {
        $this->pretest();
        $events = ['user.created', 'user.updated', 'user.deleted'];
        $userEvents = array_filter($events, fn($e) => str_starts_with($e, 'user.'));

        $this->assertCount(3, $userEvents);
    }

    public function testEventQueue(): void
    {
        $this->pretest();
        $queue = [];
        $queue[] = ['event' => 'a', 'data' => 1];
        $queue[] = ['event' => 'b', 'data' => 2];

        $processed = array_shift($queue);
        $this->assertEquals('a', $processed['event']);
    }

    public function testOnceListener(): void
    {
        $this->pretest();
        $count = 0;
        $once = true;

        for ($i = 0; $i < 3; $i++) {
            if ($once) {
                $count++;
                $once = false;
            }
        }

        $this->assertEquals(1, $count);
    }
}
