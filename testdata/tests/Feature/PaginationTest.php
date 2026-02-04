<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class PaginationTest extends TestCase
{
    private array $items;

    protected function setUp(): void
    {
        parent::setUp();
        $this->items = range(1, 100);
    }

    public function testBasicPagination(): void
    {
        $this->pretest();
        $page = 1;
        $perPage = 10;

        $paginated = $this->paginate($this->items, $page, $perPage);

        $this->assertCount(10, $paginated['data']);
        $this->assertEquals(1, $paginated['data'][0]);
    }

    public function testSecondPage(): void
    {
        $this->pretest();
        $paginated = $this->paginate($this->items, 2, 10);

        $this->assertEquals(11, $paginated['data'][0]);
        $this->assertEquals(20, $paginated['data'][9]);
    }

    public function testLastPage(): void
    {
        $this->pretest();
        $paginated = $this->paginate($this->items, 10, 10);

        $this->assertEquals(91, $paginated['data'][0]);
        $this->assertEquals(100, $paginated['data'][9]);
    }

    public function testTotalPages(): void
    {
        $this->pretest();
        $paginated = $this->paginate($this->items, 1, 10);

        $this->assertEquals(10, $paginated['last_page']);
    }

    public function testTotalCount(): void
    {
        $this->pretest();
        $paginated = $this->paginate($this->items, 1, 10);

        $this->assertEquals(100, $paginated['total']);
    }

    public function testHasMorePages(): void
    {
        $this->pretest();
        $first = $this->paginate($this->items, 1, 10);
        $last = $this->paginate($this->items, 10, 10);

        $this->assertTrue($first['has_more']);
        $this->assertFalse($last['has_more']);
    }

    public function testEmptyPage(): void
    {
        $this->pretest();
        $paginated = $this->paginate($this->items, 100, 10);

        $this->assertEmpty($paginated['data']);
    }

    public function testCustomPerPage(): void
    {
        $this->pretest();
        $paginated = $this->paginate($this->items, 1, 25);

        $this->assertCount(25, $paginated['data']);
        $this->assertEquals(4, $paginated['last_page']);
    }

    public function testFromTo(): void
    {
        $this->pretest();
        $paginated = $this->paginate($this->items, 2, 15);

        $this->assertEquals(16, $paginated['from']);
        $this->assertEquals(30, $paginated['to']);
    }

    public function testPartialLastPage(): void
    {
        $this->pretest();
        $items = range(1, 25);
        $paginated = $this->paginate($items, 3, 10);

        $this->assertCount(5, $paginated['data']);
    }

    private function paginate(array $items, int $page, int $perPage): array
    {
        $total = count($items);
        $lastPage = (int) ceil($total / $perPage);
        $offset = ($page - 1) * $perPage;

        $data = array_slice($items, $offset, $perPage);

        return [
            'data' => $data,
            'current_page' => $page,
            'per_page' => $perPage,
            'total' => $total,
            'last_page' => $lastPage,
            'from' => $offset + 1,
            'to' => $offset + count($data),
            'has_more' => $page < $lastPage,
        ];
    }
}
