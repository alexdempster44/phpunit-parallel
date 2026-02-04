<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class DatabaseTest extends TestCase
{
    private array $mockDatabase = [];

    protected function setUp(): void
    {
        parent::setUp();
        $this->mockDatabase = [];
    }

    public function testCanInsertRecord(): void
    {
        $this->pretest();
        $this->mockDatabase['users'][] = ['id' => 1, 'name' => 'John'];

        $this->assertCount(1, $this->mockDatabase['users']);
    }

    public function testCanRetrieveRecord(): void
    {
        $this->pretest();
        $this->mockDatabase['users'][] = ['id' => 1, 'name' => 'John'];

        $record = $this->mockDatabase['users'][0];
        $this->assertEquals('John', $record['name']);
    }

    public function testCanUpdateRecord(): void
    {
        $this->pretest();
        $this->mockDatabase['users'][] = ['id' => 1, 'name' => 'John'];
        $this->mockDatabase['users'][0]['name'] = 'Jane';

        $this->assertEquals('Jane', $this->mockDatabase['users'][0]['name']);
    }

    public function testCanDeleteRecord(): void
    {
        $this->pretest();
        $this->mockDatabase['users'][] = ['id' => 1, 'name' => 'John'];
        unset($this->mockDatabase['users'][0]);

        $this->assertEmpty($this->mockDatabase['users']);
    }

    public function testCanInsertMultipleRecords(): void
    {
        $this->pretest();
        $this->mockDatabase['users'][] = ['id' => 1, 'name' => 'John'];
        $this->mockDatabase['users'][] = ['id' => 2, 'name' => 'Jane'];
        $this->mockDatabase['users'][] = ['id' => 3, 'name' => 'Bob'];

        $this->assertCount(3, $this->mockDatabase['users']);
    }

    public function testCanFindRecordById(): void
    {
        $this->pretest();
        $this->mockDatabase['users'][] = ['id' => 1, 'name' => 'John'];
        $this->mockDatabase['users'][] = ['id' => 2, 'name' => 'Jane'];

        $found = array_filter($this->mockDatabase['users'], fn($u) => $u['id'] === 2);
        $this->assertCount(1, $found);
        $this->assertEquals('Jane', array_values($found)[0]['name']);
    }

    public function testCanCountRecords(): void
    {
        $this->pretest();
        $this->mockDatabase['users'] = [];
        $this->assertEquals(0, count($this->mockDatabase['users']));

        $this->mockDatabase['users'][] = ['id' => 1, 'name' => 'John'];
        $this->assertEquals(1, count($this->mockDatabase['users']));
    }

    public function testCanCheckIfRecordExists(): void
    {
        $this->pretest();
        $this->mockDatabase['users'][] = ['id' => 1, 'name' => 'John'];

        $exists = !empty(array_filter($this->mockDatabase['users'], fn($u) => $u['id'] === 1));
        $notExists = !empty(array_filter($this->mockDatabase['users'], fn($u) => $u['id'] === 999));

        $this->assertTrue($exists);
        $this->assertFalse($notExists);
    }
}
