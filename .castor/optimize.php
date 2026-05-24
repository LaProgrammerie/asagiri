<?php

namespace optimize;

use Castor\Attribute\AsTask;

use function Castor\io;
use function Castor\variable;
use function docker\docker_compose_run;

function read_package_json(): ?array
{
    $path = variable('root_dir') . '/application/package.json';
    if (!is_file($path)) {
        return null;
    }
    $json = json_decode((string) file_get_contents($path), true);

    return is_array($json) ? $json : null;
}

function has_script(string $name): bool
{
    $package = read_package_json();

    return is_array($package) && isset($package['scripts'][$name]) && is_string($package['scripts'][$name]);
}

#[AsTask(description: 'Optimize images when script exists', aliases: ['images'])]
function images(): void
{
    if (!has_script('optimize:images')) {
        io()->note('Script npm optimize:images missing, skipping.');

        return;
    }

    io()->title('Image optimization');
    docker_compose_run('npm run optimize:images', workDir: '/var/www/application');
}

#[AsTask(description: 'Generate Open Graph image when script exists', aliases: ['og-image'])]
function og_image(): void
{
    if (!has_script('optimize:og-image')) {
        io()->note('Script npm optimize:og-image missing, skipping.');

        return;
    }

    io()->title('Generating OG image');
    docker_compose_run('npm run optimize:og-image', workDir: '/var/www/application');
}

#[AsTask(description: 'Measure Web Vitals when script exists', aliases: ['web-vitals'])]
function web_vitals(): void
{
    if (!has_script('optimize:web-vitals')) {
        io()->note('Script npm optimize:web-vitals missing, skipping.');

        return;
    }

    io()->title('Measuring Web Vitals');
    docker_compose_run('npm run optimize:web-vitals', workDir: '/var/www/application');
}

#[AsTask(description: 'Run all optimization tasks', aliases: ['all'])]
function all(): void
{
    images();
    og_image();
    web_vitals();
}
