<?php

declare(strict_types=1);

namespace App\Service;

use Symfony\Component\DependencyInjection\ParameterBag\ParameterBagInterface;

final class AssetPathService
{
    public function __construct(private readonly ParameterBagInterface $parameterBag)
    {
    }

    public function baseUrl(string $env = 'prod'): string
    {
        return match ($env) {
            'dev' => $this->bagString('app.assets.dev.base_url'),
            'cdn' => $this->bagString('app.assets.cdn.base_url'),
            default => $this->bagString('app.assets.prod.base_url'),
        };
    }

    private function bagString(string $key): string
    {
        if (!$this->parameterBag->has($key)) {
            return '';
        }

        $value = $this->parameterBag->get($key);

        return is_string($value) ? $value : '';
    }
}
