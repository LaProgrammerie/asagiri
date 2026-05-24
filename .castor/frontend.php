<?php

namespace frontend;

use Castor\Attribute\AsTask;

use function Castor\io;
use function Castor\variable;
use function docker\docker_compose_run;

function has_package_json(): bool
{
    return is_file(variable('root_dir') . '/application/package.json');
}

#[AsTask(description: 'Build production frontend assets (if Node app is present)', namespace: 'app', name: 'assets:prod')]
function assets_prod(): void
{
    if (!has_package_json()) {
        io()->note('No application/package.json, skipping app:assets:prod.');

        return;
    }

    io()->title('Building production frontend assets');
    docker_compose_run('npm run build', workDir: '/var/www/application');
    io()->success('Assets build completed');
}

#[AsTask(description: 'Build frontend assets in development mode', namespace: 'app', name: 'assets:dev')]
function assets_dev(): void
{
    if (!has_package_json()) {
        io()->note('No application/package.json, skipping app:assets:dev.');

        return;
    }

    io()->title('Building frontend assets (dev)');
    docker_compose_run('npm run dev', workDir: '/var/www/application');
}

#[AsTask(description: 'Watch frontend assets', namespace: 'app', name: 'assets:watch')]
function assets_watch(): void
{
    if (!has_package_json()) {
        io()->note('No application/package.json, skipping app:assets:watch.');

        return;
    }

    io()->title('Watching frontend assets');
    docker_compose_run('npm run watch', workDir: '/var/www/application');
}
