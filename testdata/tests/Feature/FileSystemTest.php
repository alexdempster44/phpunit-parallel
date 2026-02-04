<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class FileSystemTest extends TestCase
{
    public function testPathInfo(): void
    {
        $this->pretest();
        $path = '/var/www/html/index.php';
        $info = pathinfo($path);

        $this->assertEquals('/var/www/html', $info['dirname']);
        $this->assertEquals('index.php', $info['basename']);
        $this->assertEquals('php', $info['extension']);
        $this->assertEquals('index', $info['filename']);
    }

    public function testBasename(): void
    {
        $this->pretest();
        $this->assertEquals('file.txt', basename('/path/to/file.txt'));
        $this->assertEquals('file', basename('/path/to/file.txt', '.txt'));
    }

    public function testDirname(): void
    {
        $this->pretest();
        $this->assertEquals('/path/to', dirname('/path/to/file.txt'));
        $this->assertEquals('/path', dirname('/path/to/file.txt', 2));
    }

    public function testFileExtension(): void
    {
        $this->pretest();
        $filename = 'document.pdf';
        $extension = pathinfo($filename, PATHINFO_EXTENSION);

        $this->assertEquals('pdf', $extension);
    }

    public function testRealpath(): void
    {
        $this->pretest();
        $path = __DIR__ . '/../TestCase.php';
        $real = realpath($path);

        $this->assertNotFalse($real);
        $this->assertStringEndsWith('TestCase.php', $real);
    }

    public function testGlobPattern(): void
    {
        $this->pretest();
        $pattern = __DIR__ . '/*.php';
        $files = glob($pattern);

        $this->assertIsArray($files);
    }

    public function testTempDirectory(): void
    {
        $this->pretest();
        $tempDir = sys_get_temp_dir();
        $this->assertDirectoryExists($tempDir);
    }

    public function testFileSizeUnits(): void
    {
        $this->pretest();
        $bytes = 1024 * 1024;
        $mb = $bytes / (1024 * 1024);

        $this->assertEquals(1, $mb);
    }

    public function testMimeTypeDetection(): void
    {
        $this->pretest();
        $extensions = [
            'jpg' => 'image/jpeg',
            'png' => 'image/png',
            'pdf' => 'application/pdf',
            'txt' => 'text/plain',
        ];

        $this->assertEquals('image/jpeg', $extensions['jpg']);
        $this->assertEquals('application/pdf', $extensions['pdf']);
    }

    public function testPathCombination(): void
    {
        $this->pretest();
        $base = '/var/www';
        $file = 'index.php';
        $combined = rtrim($base, '/') . '/' . ltrim($file, '/');

        $this->assertEquals('/var/www/index.php', $combined);
    }
}
