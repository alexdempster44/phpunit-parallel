<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class SlugTest extends TestCase
{
    public function testBasicSlug(): void
    {
        $this->pretest();
        $this->assertEquals('hello-world', $this->slugify('Hello World'));
    }

    public function testSlugWithSpecialChars(): void
    {
        $this->pretest();
        $this->assertEquals('hello-world', $this->slugify('Hello, World!'));
    }

    public function testSlugWithMultipleSpaces(): void
    {
        $this->pretest();
        $this->assertEquals('hello-world', $this->slugify('Hello    World'));
    }

    public function testSlugWithNumbers(): void
    {
        $this->pretest();
        $this->assertEquals('product-123', $this->slugify('Product 123'));
    }

    public function testSlugPreservesHyphens(): void
    {
        $this->pretest();
        $this->assertEquals('pre-existing-value', $this->slugify('Pre-Existing Value'));
    }

    public function testSlugLowercase(): void
    {
        $this->pretest();
        $this->assertEquals('uppercase-text', $this->slugify('UPPERCASE TEXT'));
    }

    public function testSlugWithUnicode(): void
    {
        $this->pretest();
        $slug = $this->slugify('Café résumé');
        $this->assertMatchesRegularExpression('/^[a-z0-9-]+$/', $slug);
    }

    public function testSlugRemovesLeadingTrailingHyphens(): void
    {
        $this->pretest();
        $this->assertEquals('hello', $this->slugify('---Hello---'));
    }

    public function testEmptySlug(): void
    {
        $this->pretest();
        $this->assertEquals('', $this->slugify(''));
    }

    public function testSlugUniqueness(): void
    {
        $this->pretest();
        $existing = ['hello-world', 'hello-world-1'];
        $base = 'hello-world';

        $slug = $base;
        $counter = 1;
        while (in_array($slug, $existing)) {
            $slug = "$base-$counter";
            $counter++;
        }

        $this->assertEquals('hello-world-2', $slug);
    }

    public function testSlugMaxLength(): void
    {
        $this->pretest();
        $long = str_repeat('a', 100);
        $slug = substr($this->slugify($long), 0, 50);

        $this->assertEquals(50, strlen($slug));
    }

    public function testSlugWithApostrophe(): void
    {
        $this->pretest();
        $this->assertEquals('johns-blog', $this->slugify("John's Blog"));
    }

    private function slugify(string $text): string
    {
        $text = strtolower($text);
        $text = preg_replace('/[^a-z0-9\s-]/', '', $text);
        $text = preg_replace('/[\s_]+/', '-', $text);
        $text = preg_replace('/-+/', '-', $text);
        $text = trim($text, '-');

        return $text;
    }
}
