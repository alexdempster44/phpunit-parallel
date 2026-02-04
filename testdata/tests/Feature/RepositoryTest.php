<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class RepositoryTest extends TestCase
{
    private array $data = [];
    private int $nextId = 1;

    protected function setUp(): void
    {
        parent::setUp();
        $this->data = [];
        $this->nextId = 1;
    }

    public function testCreate(): void
    {
        $this->pretest();
        $entity = $this->create(['name' => 'John']);

        $this->assertEquals(1, $entity['id']);
        $this->assertEquals('John', $entity['name']);
    }

    public function testFind(): void
    {
        $this->pretest();
        $this->create(['name' => 'John']);
        $this->create(['name' => 'Jane']);

        $found = $this->find(2);
        $this->assertEquals('Jane', $found['name']);
    }

    public function testFindOrFail(): void
    {
        $this->pretest();
        $this->create(['name' => 'John']);

        $this->expectException(\RuntimeException::class);
        $this->findOrFail(999);
    }

    public function testAll(): void
    {
        $this->pretest();
        $this->create(['name' => 'A']);
        $this->create(['name' => 'B']);
        $this->create(['name' => 'C']);

        $this->assertCount(3, $this->all());
    }

    public function testUpdate(): void
    {
        $this->pretest();
        $entity = $this->create(['name' => 'John']);
        $this->update($entity['id'], ['name' => 'Jane']);

        $updated = $this->find($entity['id']);
        $this->assertEquals('Jane', $updated['name']);
    }

    public function testDelete(): void
    {
        $this->pretest();
        $entity = $this->create(['name' => 'John']);
        $this->delete($entity['id']);

        $this->assertNull($this->find($entity['id']));
    }

    public function testWhere(): void
    {
        $this->pretest();
        $this->create(['name' => 'John', 'active' => true]);
        $this->create(['name' => 'Jane', 'active' => false]);
        $this->create(['name' => 'Bob', 'active' => true]);

        $active = $this->where('active', true);
        $this->assertCount(2, $active);
    }

    public function testFirstWhere(): void
    {
        $this->pretest();
        $this->create(['name' => 'John', 'role' => 'admin']);
        $this->create(['name' => 'Jane', 'role' => 'user']);

        $admin = $this->firstWhere('role', 'admin');
        $this->assertEquals('John', $admin['name']);
    }

    public function testCount(): void
    {
        $this->pretest();
        $this->create(['name' => 'A']);
        $this->create(['name' => 'B']);

        $this->assertEquals(2, count($this->data));
    }

    private function create(array $attributes): array
    {
        $entity = array_merge(['id' => $this->nextId++], $attributes);
        $this->data[$entity['id']] = $entity;
        return $entity;
    }

    private function find(int $id): ?array
    {
        return $this->data[$id] ?? null;
    }

    private function findOrFail(int $id): array
    {
        return $this->find($id) ?? throw new \RuntimeException('Not found');
    }

    private function all(): array
    {
        return array_values($this->data);
    }

    private function update(int $id, array $attributes): void
    {
        if (isset($this->data[$id])) {
            $this->data[$id] = array_merge($this->data[$id], $attributes);
        }
    }

    private function delete(int $id): void
    {
        unset($this->data[$id]);
    }

    private function where(string $field, mixed $value): array
    {
        return array_values(array_filter($this->data, fn($e) => ($e[$field] ?? null) === $value));
    }

    private function firstWhere(string $field, mixed $value): ?array
    {
        $results = $this->where($field, $value);
        return $results[0] ?? null;
    }
}
