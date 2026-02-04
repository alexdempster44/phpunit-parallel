<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class SerializationTest extends TestCase
{
    public function testSerializeArray(): void
    {
        $this->pretest();
        $data = ['name' => 'John', 'age' => 30];
        $serialized = serialize($data);
        $unserialized = unserialize($serialized);

        $this->assertEquals($data, $unserialized);
    }

    public function testSerializeObject(): void
    {
        $this->pretest();
        $obj = new \stdClass();
        $obj->value = 42;

        $serialized = serialize($obj);
        $unserialized = unserialize($serialized);

        $this->assertEquals(42, $unserialized->value);
    }

    public function testJsonSerialize(): void
    {
        $this->pretest();
        $data = ['items' => [1, 2, 3]];
        $json = json_encode($data);
        $decoded = json_decode($json, true);

        $this->assertEquals($data, $decoded);
    }

    public function testBase64Encode(): void
    {
        $this->pretest();
        $data = 'Hello, World!';
        $encoded = base64_encode($data);
        $decoded = base64_decode($encoded);

        $this->assertEquals($data, $decoded);
    }

    public function testUrlEncode(): void
    {
        $this->pretest();
        $data = 'hello world&foo=bar';
        $encoded = urlencode($data);
        $decoded = urldecode($encoded);

        $this->assertEquals($data, $decoded);
        $this->assertStringContainsString('+', $encoded);
    }

    public function testRawUrlEncode(): void
    {
        $this->pretest();
        $data = 'hello world';
        $encoded = rawurlencode($data);

        $this->assertEquals('hello%20world', $encoded);
    }

    public function testHttpBuildQuery(): void
    {
        $this->pretest();
        $params = ['name' => 'John', 'age' => 30];
        $query = http_build_query($params);

        $this->assertEquals('name=John&age=30', $query);
    }

    public function testParseStr(): void
    {
        $this->pretest();
        $query = 'name=John&age=30';
        parse_str($query, $result);

        $this->assertEquals('John', $result['name']);
        $this->assertEquals('30', $result['age']);
    }

    public function testPackUnpack(): void
    {
        $this->pretest();
        $packed = pack('N', 12345);
        $unpacked = unpack('N', $packed);

        $this->assertEquals(12345, $unpacked[1]);
    }

    public function testIgbinaryAlternative(): void
    {
        $this->pretest();
        $data = ['key' => 'value'];
        $serialized = serialize($data);

        $this->assertIsString($serialized);
        $this->assertStringStartsWith('a:', $serialized);
    }
}
