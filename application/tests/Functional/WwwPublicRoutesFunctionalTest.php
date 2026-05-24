<?php

declare(strict_types=1);

namespace App\Tests\Functional;

use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;

final class WwwPublicRoutesFunctionalTest extends WebTestCase
{
    public function testRootRouteReturnsSuccessfulResponse(): void
    {
        $client = static::createClient();
        $client->request('GET', '/', server: ['HTTP_HOST' => 'app.test']);

        if (500 === $client->getResponse()->getStatusCode()) {
            usleep(350000);
            $client->request('GET', '/', server: ['HTTP_HOST' => 'app.test']);
        }

        self::assertTrue(
            $client->getResponse()->isSuccessful(),
            sprintf('Unexpected status code: %d', $client->getResponse()->getStatusCode())
        );
    }
}
