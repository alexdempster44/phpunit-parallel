<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;
use SplQueue;

class QueueTest extends TestCase
{
    public function testEnqueueDequeue(): void
    {
        $this->pretest();
        $queue = new SplQueue();
        $queue->enqueue('first');
        $queue->enqueue('second');

        $this->assertEquals('first', $queue->dequeue());
        $this->assertEquals('second', $queue->dequeue());
    }

    public function testFifoOrder(): void
    {
        $this->pretest();
        $queue = new SplQueue();
        $queue->enqueue(1);
        $queue->enqueue(2);
        $queue->enqueue(3);

        $result = [];
        while (!$queue->isEmpty()) {
            $result[] = $queue->dequeue();
        }

        $this->assertEquals([1, 2, 3], $result);
    }

    public function testIsEmpty(): void
    {
        $this->pretest();
        $queue = new SplQueue();
        $this->assertTrue($queue->isEmpty());

        $queue->enqueue('item');
        $this->assertFalse($queue->isEmpty());
    }

    public function testCount(): void
    {
        $this->pretest();
        $queue = new SplQueue();
        $this->assertEquals(0, $queue->count());

        $queue->enqueue('a');
        $queue->enqueue('b');
        $this->assertEquals(2, $queue->count());
    }

    public function testPeek(): void
    {
        $this->pretest();
        $queue = new SplQueue();
        $queue->enqueue('first');
        $queue->enqueue('second');

        $this->assertEquals('first', $queue->bottom());
        $this->assertEquals(2, $queue->count());
    }

    public function testArrayAsQueue(): void
    {
        $this->pretest();
        $queue = [];
        $queue[] = 'first';
        $queue[] = 'second';

        $this->assertEquals('first', array_shift($queue));
        $this->assertEquals('second', array_shift($queue));
    }

    public function testPrioritySimulation(): void
    {
        $this->pretest();
        $queue = [
            ['priority' => 1, 'data' => 'low'],
            ['priority' => 3, 'data' => 'high'],
            ['priority' => 2, 'data' => 'medium'],
        ];

        usort($queue, fn($a, $b) => $b['priority'] <=> $a['priority']);

        $this->assertEquals('high', $queue[0]['data']);
    }

    public function testBatchProcessing(): void
    {
        $this->pretest();
        $queue = range(1, 10);
        $batchSize = 3;
        $batches = array_chunk($queue, $batchSize);

        $this->assertCount(4, $batches);
        $this->assertEquals([1, 2, 3], $batches[0]);
    }

    public function testJobQueue(): void
    {
        $this->pretest();
        $jobs = [];
        $jobs[] = ['type' => 'email', 'to' => 'user@example.com'];
        $jobs[] = ['type' => 'sms', 'to' => '+1234567890'];

        $job = array_shift($jobs);
        $this->assertEquals('email', $job['type']);
    }
}
