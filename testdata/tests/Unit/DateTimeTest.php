<?php

declare(strict_types=1);

namespace Tests\Unit;

use DateTime;
use DateTimeImmutable;
use DateInterval;
use Tests\TestCase;

class DateTimeTest extends TestCase
{
    public function testCreateDateTime(): void
    {
        $this->pretest();
        $date = new DateTime('2024-01-15');
        $this->assertEquals('2024-01-15', $date->format('Y-m-d'));
    }

    public function testDateTimeImmutable(): void
    {
        $this->pretest();
        $date = new DateTimeImmutable('2024-01-15');
        $newDate = $date->modify('+1 day');

        $this->assertEquals('2024-01-15', $date->format('Y-m-d'));
        $this->assertEquals('2024-01-16', $newDate->format('Y-m-d'));
    }

    public function testDateDiff(): void
    {
        $this->pretest();
        $date1 = new DateTime('2024-01-01');
        $date2 = new DateTime('2024-01-10');
        $diff = $date1->diff($date2);

        $this->assertEquals(9, $diff->days);
    }

    public function testAddInterval(): void
    {
        $this->pretest();
        $date = new DateTime('2024-01-15');
        $date->add(new DateInterval('P1M'));

        $this->assertEquals('2024-02-15', $date->format('Y-m-d'));
    }

    public function testTimestamp(): void
    {
        $this->pretest();
        $date = new DateTime('2024-01-01 00:00:00', new \DateTimeZone('UTC'));
        $timestamp = $date->getTimestamp();

        $this->assertIsInt($timestamp);
    }

    public function testDateFormat(): void
    {
        $this->pretest();
        $date = new DateTime('2024-03-15 14:30:00');

        $this->assertEquals('March', $date->format('F'));
        $this->assertEquals('Fri', $date->format('D'));
        $this->assertEquals('14:30', $date->format('H:i'));
    }

    public function testIsWeekend(): void
    {
        $this->pretest();
        $saturday = new DateTime('2024-01-13');
        $monday = new DateTime('2024-01-15');

        $this->assertTrue(in_array($saturday->format('N'), ['6', '7']));
        $this->assertFalse(in_array($monday->format('N'), ['6', '7']));
    }

    public function testLeapYear(): void
    {
        $this->pretest();
        $leapYear = new DateTime('2024-02-29');
        $this->assertEquals('29', $leapYear->format('d'));
    }
}
