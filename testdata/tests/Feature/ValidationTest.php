<?php

declare(strict_types=1);

namespace Tests\Feature;

use DateTime;
use Tests\TestCase;

class ValidationTest extends TestCase
{
    public function testEmailValidation(): void
    {
        $this->pretest();
        $validEmail = 'user@example.com';
        $invalidEmail = 'not-an-email';

        $this->assertTrue(filter_var($validEmail, FILTER_VALIDATE_EMAIL) !== false);
        $this->assertFalse(filter_var($invalidEmail, FILTER_VALIDATE_EMAIL) !== false);
    }

    public function testRequiredFieldValidation(): void
    {
        $this->pretest();
        $data = ['name' => 'John', 'email' => ''];
        $required = ['name', 'email'];
        $errors = [];

        foreach ($required as $field) {
            if (empty($data[$field])) {
                $errors[] = "$field is required";
            }
        }

        $this->assertContains('email is required', $errors);
        $this->assertNotContains('name is required', $errors);
    }

    public function testNumericValidation(): void
    {
        $this->pretest();
        $this->assertTrue(is_numeric('123'));
        $this->assertTrue(is_numeric('12.34'));
        $this->assertFalse(is_numeric('abc'));
    }

    public function testUrlValidation(): void
    {
        $this->pretest();
        $validUrl = 'https://example.com';
        $invalidUrl = 'not-a-url';

        $this->assertTrue(filter_var($validUrl, FILTER_VALIDATE_URL) !== false);
        $this->assertFalse(filter_var($invalidUrl, FILTER_VALIDATE_URL) !== false);
    }

    public function testIpAddressValidation(): void
    {
        $this->pretest();
        $this->assertTrue(filter_var('192.168.1.1', FILTER_VALIDATE_IP) !== false);
        $this->assertTrue(filter_var('::1', FILTER_VALIDATE_IP) !== false);
        $this->assertFalse(filter_var('999.999.999.999', FILTER_VALIDATE_IP) !== false);
    }

    public function testIntegerValidation(): void
    {
        $this->pretest();
        $this->assertTrue(filter_var('42', FILTER_VALIDATE_INT) !== false);
        $this->assertFalse(filter_var('42.5', FILTER_VALIDATE_INT) !== false);
        $this->assertFalse(filter_var('abc', FILTER_VALIDATE_INT) !== false);
    }

    public function testBooleanValidation(): void
    {
        $this->pretest();
        $this->assertTrue(filter_var('true', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE) === true);
        $this->assertTrue(filter_var('1', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE) === true);
        $this->assertTrue(filter_var('false', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE) === false);
        $this->assertNull(filter_var('invalid', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE));
    }

    public function testMinLengthValidation(): void
    {
        $this->pretest();
        $password = 'abc';
        $minLength = 8;

        $this->assertFalse(strlen($password) >= $minLength);
        $this->assertTrue(strlen('password123') >= $minLength);
    }

    public function testMaxLengthValidation(): void
    {
        $this->pretest();
        $username = 'verylongusernamethatexceedslimit';
        $maxLength = 20;

        $this->assertFalse(strlen($username) <= $maxLength);
        $this->assertTrue(strlen('shortname') <= $maxLength);
    }

    public function testRegexValidation(): void
    {
        $this->pretest();
        $pattern = '/^[a-z0-9_]+$/';

        $this->assertTrue(preg_match($pattern, 'valid_username_123') === 1);
        $this->assertFalse(preg_match($pattern, 'Invalid Username!') === 1);
    }

    public function testDateValidation(): void
    {
        $this->pretest();
        $validDate = '2024-01-15';
        $invalidDate = '2024-13-45';

        $this->assertTrue(strtotime($validDate) !== false);
        $d = DateTime::createFromFormat('Y-m-d', $invalidDate);
        $this->assertFalse($d && $d->format('Y-m-d') === $invalidDate);
    }
}
