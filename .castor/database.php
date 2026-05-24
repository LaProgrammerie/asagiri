<?php

use Castor\Attribute\AsTask;

use function Castor\context;
use function Castor\io;
use function Castor\variable;
use function docker\docker_compose;
use function docker\docker_compose_run;

#[AsTask(description: 'Connect to the PostgreSQL database', name: 'db:client', aliases: ['postgres', 'pg'])]
function postgres_client(): void
{
    io()->title('Connecting to the PostgreSQL database');

    docker_compose(['exec', 'postgres', 'psql', '-U', 'app', 'app'], context()->toInteractive());
}

function symfony_console_exists(): bool
{
    return is_file(variable('root_dir') . '/application/bin/console');
}

#[AsTask(description: 'Wait until PostgreSQL accepts TCP', name: 'db:wait')]
function wait_for_postgres(int $maxAttempts = 45, int $sleepSeconds = 2): void
{
    io()->section('Waiting for PostgreSQL to accept TCP connections');
    $loop = sprintf(
        'for i in $(seq 1 %d); do php -r \'$f=@fsockopen("postgres", 5432); if ($f) { fclose($f); exit(0); } exit(1);\' && exit 0; sleep %d; done; exit 1',
        $maxAttempts,
        $sleepSeconds
    );
    docker_compose_run($loop, noDeps: false);
}

#[AsTask(description: 'Database setup for CI/local (idempotent)', name: 'db:setup')]
function database_setup(): void
{
    wait_for_postgres();
    if (!symfony_console_exists()) {
        io()->note('No Symfony console found, db:setup skipped.');

        return;
    }

    docker_compose_run('bin/console doctrine:database:create --if-not-exists', noDeps: false, workDir: '/var/www/application');
    docker_compose_run(
        'if [ -d migrations ] && [ "$(ls -A migrations/*.php 2>/dev/null | wc -l)" -gt 0 ]; then bin/console doctrine:migrations:migrate -n; else bin/console doctrine:schema:update --force; fi',
        noDeps: false,
        workDir: '/var/www/application'
    );
}

#[AsTask(description: 'Load fixtures on APP_ENV=test when available', name: 'db:fixtures-test', aliases: ['fixtures:app_test'])]
function fixtures_test(): void
{
    wait_for_postgres();
    if (!symfony_console_exists()) {
        io()->note('No Symfony console found, db:fixtures-test skipped.');

        return;
    }

    docker_compose_run('bin/console doctrine:fixtures:load -n --env=test', noDeps: false, workDir: '/var/www/application');
}
