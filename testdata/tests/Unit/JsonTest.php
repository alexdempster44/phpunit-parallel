<?php

declare(strict_types=1);

namespace Tests\Unit;

use Tests\TestCase;

class JsonTest extends TestCase
{
    public function testJsonEncode(): void
    {
        $this->pretest();
        $data = ['name' => 'John', 'age' => 30];
        $json = json_encode($data);
        $this->assertEquals('{"name":"John","age":30}', $json);
    }

    public function testJsonDecode(): void
    {
        $this->pretest();
        $json = '{"name":"John","age":30}';
        $data = json_decode($json, true);
        $this->assertEquals('John', $data['name']);
        $this->assertEquals(30, $data['age']);
    }

    public function testJsonDecodeAsObject(): void
    {
        $this->pretest();
        $json = '{"name":"John","age":30}';
        $data = json_decode($json);
        $this->assertEquals('John', $data->name);
        $this->assertEquals(30, $data->age);
    }

    public function testJsonEncodeArray(): void
    {
        $this->pretest();
        $data = [1, 2, 3];
        $json = json_encode($data);
        $this->assertEquals('[1,2,3]', $json);
    }

    public function testJsonEncodePretty(): void
    {
        $this->pretest();
        $data = ['a' => 1];
        $json = json_encode($data, JSON_PRETTY_PRINT);
        $this->assertStringContainsString("\n", $json);
    }

    public function testJsonEncodeUnicode(): void
    {
        $this->pretest();
        $data = ['text' => 'Hello 世界'];
        $json = json_encode($data, JSON_UNESCAPED_UNICODE);
        $this->assertStringContainsString('世界', $json);
    }

    public function testJsonDecodeNestedArray(): void
    {
        $this->pretest();
        $json = '{"user":{"name":"John","address":{"city":"NYC"}}}';
        $data = json_decode($json, true);
        $this->assertEquals('NYC', $data['user']['address']['city']);
    }

    public function testJsonEncodeNull(): void
    {
        $this->pretest();
        $this->assertEquals('null', json_encode(null));
    }

    public function testJsonEncodeBool(): void
    {
        $this->pretest();
        $this->assertEquals('true', json_encode(true));
        $this->assertEquals('false', json_encode(false));
    }

    public function testJsonLastError(): void
    {
        $this->pretest();
        json_decode('{"valid": "json"}');
        $this->assertEquals(JSON_ERROR_NONE, json_last_error());
    }
}
