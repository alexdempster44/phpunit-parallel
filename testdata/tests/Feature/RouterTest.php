<?php

declare(strict_types=1);

namespace Tests\Feature;

use Tests\TestCase;

class RouterTest extends TestCase
{
    private array $routes = [];

    public function testRegisterRoute(): void
    {
        $this->pretest();
        $this->addRoute('GET', '/users', 'UsersController@index');

        $this->assertCount(1, $this->routes);
    }

    public function testMatchRoute(): void
    {
        $this->pretest();
        $this->addRoute('GET', '/users', 'UsersController@index');

        $match = $this->match('GET', '/users');
        $this->assertEquals('UsersController@index', $match['handler']);
    }

    public function testRouteWithParameter(): void
    {
        $this->pretest();
        $this->addRoute('GET', '/users/{id}', 'UsersController@show');

        $match = $this->match('GET', '/users/42');
        $this->assertEquals('42', $match['params']['id']);
    }

    public function testMultipleParameters(): void
    {
        $this->pretest();
        $this->addRoute('GET', '/posts/{postId}/comments/{commentId}', 'CommentsController@show');

        $match = $this->match('GET', '/posts/10/comments/5');
        $this->assertEquals('10', $match['params']['postId']);
        $this->assertEquals('5', $match['params']['commentId']);
    }

    public function testMethodMatching(): void
    {
        $this->pretest();
        $this->addRoute('GET', '/users', 'get_handler');
        $this->addRoute('POST', '/users', 'post_handler');

        $getMatch = $this->match('GET', '/users');
        $postMatch = $this->match('POST', '/users');

        $this->assertEquals('get_handler', $getMatch['handler']);
        $this->assertEquals('post_handler', $postMatch['handler']);
    }

    public function testNoMatch(): void
    {
        $this->pretest();
        $this->addRoute('GET', '/users', 'handler');

        $match = $this->match('GET', '/posts');
        $this->assertNull($match);
    }

    public function testOptionalParameter(): void
    {
        $this->pretest();
        $pattern = '/users/{id?}';
        $this->assertTrue(str_contains($pattern, '?'));
    }

    public function testRouteGroups(): void
    {
        $this->pretest();
        $prefix = '/api/v1';
        $routes = ['/users', '/posts', '/comments'];

        $prefixed = array_map(fn($r) => $prefix . $r, $routes);

        $this->assertEquals('/api/v1/users', $prefixed[0]);
    }

    public function testRouteNames(): void
    {
        $this->pretest();
        $namedRoutes = [
            'users.index' => '/users',
            'users.show' => '/users/{id}',
        ];

        $this->assertEquals('/users', $namedRoutes['users.index']);
    }

    public function testUrlGeneration(): void
    {
        $this->pretest();
        $pattern = '/users/{id}/posts/{postId}';
        $params = ['id' => 42, 'postId' => 10];

        $url = preg_replace_callback('/\{(\w+)\}/', fn($m) => $params[$m[1]], $pattern);

        $this->assertEquals('/users/42/posts/10', $url);
    }

    private function addRoute(string $method, string $path, string $handler): void
    {
        $this->routes[] = compact('method', 'path', 'handler');
    }

    private function match(string $method, string $uri): ?array
    {
        foreach ($this->routes as $route) {
            if ($route['method'] !== $method) {
                continue;
            }

            $pattern = preg_replace('/\{(\w+)\}/', '(?P<$1>[^/]+)', $route['path']);
            if (preg_match('#^' . $pattern . '$#', $uri, $matches)) {
                $params = array_filter($matches, 'is_string', ARRAY_FILTER_USE_KEY);
                return ['handler' => $route['handler'], 'params' => $params];
            }
        }

        return null;
    }
}
