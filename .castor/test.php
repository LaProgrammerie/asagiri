<?php

namespace test;

use Castor\Attribute\AsTask;

use function Castor\io;
use function Castor\variable;
use function docker\docker_compose_run;

function has_script(string $script): bool
{
    $path = variable('root_dir') . '/application/package.json';
    if (!is_file($path)) {
        return false;
    }
    $json = json_decode((string) file_get_contents($path), true);

    return is_array($json) && isset($json['scripts'][$script]) && is_string($json['scripts'][$script]);
}

function run_npm_script_if_exists(string $script): void
{
    if (!has_script($script)) {
        io()->note(sprintf('Script npm %s missing, skipping.', $script));

        return;
    }
    docker_compose_run('npm run ' . $script, workDir: '/var/www/application');
}

#[AsTask(description: 'Run accessibility checks when configured', aliases: ['accessibility', 'a11y'])]
function accessibility(): void
{
    io()->title('Accessibility checks');
    run_npm_script_if_exists('test:accessibility');
}

#[AsTask(description: 'Run mobile UX checks when configured', aliases: ['mobile-ux'])]
function mobile_ux(): void
{
    io()->title('Mobile UX checks');
    run_npm_script_if_exists('test:mobile-ux');
}

#[AsTask(description: 'Run optimization checks when configured', aliases: ['optimizations'])]
function optimizations(): void
{
    io()->title('Optimization checks');
    run_npm_script_if_exists('test:optimizations');
}

#[AsTask(description: 'Run all optional frontend quality checks', aliases: ['all-optimizations'])]
function all_optimizations(): void
{
    accessibility();
    mobile_ux();
    optimizations();
}
