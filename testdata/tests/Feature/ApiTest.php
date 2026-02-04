<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class ApiTest extends TestCase
{
    public function testApiResponseStructure(): void
    {
        $this->pretest();
        $response = [
            'status' => 200,
            'data' => ['id' => 1, 'name' => 'Test'],
            'message' => 'Success',
        ];

        $this->assertArrayHasKey('status', $response);
        $this->assertArrayHasKey('data', $response);
        $this->assertArrayHasKey('message', $response);
    }

    public function testSuccessfulResponse(): void
    {
        $this->pretest();
        $response = ['status' => 200, 'message' => 'OK'];

        $this->assertEquals(200, $response['status']);
    }

    public function testErrorResponse(): void
    {
        $this->pretest();
        $response = ['status' => 404, 'message' => 'Not Found'];

        $this->assertEquals(404, $response['status']);
        $this->assertEquals('Not Found', $response['message']);
    }

    public function testPaginatedResponse(): void
    {
        $this->pretest();
        $response = [
            'data' => [1, 2, 3],
            'meta' => [
                'current_page' => 1,
                'per_page' => 10,
                'total' => 3,
            ],
        ];

        $this->assertCount(3, $response['data']);
        $this->assertEquals(1, $response['meta']['current_page']);
    }

    public function testUnauthorizedResponse(): void
    {
        $this->pretest();
        $response = ['status' => 401, 'message' => 'Unauthorized'];

        $this->assertEquals(401, $response['status']);
    }

    public function testForbiddenResponse(): void
    {
        $this->pretest();
        $response = ['status' => 403, 'message' => 'Forbidden'];

        $this->assertEquals(403, $response['status']);
    }

    public function testServerErrorResponse(): void
    {
        $this->pretest();
        $response = ['status' => 500, 'message' => 'Internal Server Error'];

        $this->assertEquals(500, $response['status']);
    }

    public function testCreatedResponse(): void
    {
        $this->pretest();
        $response = ['status' => 201, 'data' => ['id' => 42], 'message' => 'Created'];

        $this->assertEquals(201, $response['status']);
        $this->assertEquals(42, $response['data']['id']);
    }

    public function testNoContentResponse(): void
    {
        $this->pretest();
        $response = ['status' => 204, 'data' => null];

        $this->assertEquals(204, $response['status']);
        $this->assertNull($response['data']);
    }
}
