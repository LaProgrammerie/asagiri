<?php

namespace cypress;

use Castor\Attribute\AsTask;

use function Castor\io;
use function Castor\variable;
use function docker\docker_compose;

function e2e_dir_exists(): bool
{
    return is_dir(variable('root_dir') . '/application/tests/E2E');
}

#[AsTask(description: 'Runs all Cypress tests when E2E scaffold is present', namespace: 'cypress', aliases: ['test:all'])]
function test_all(): void
{
    if (!e2e_dir_exists()) {
        io()->note('No application/tests/E2E directory, Cypress suite skipped.');

        return;
    }

    io()->title('Running Cypress tests');
    docker_compose(['run', '--rm', 'cypress'], profiles: ['default', 'test']);
}

#[AsTask(description: 'Opens Cypress in interactive mode', namespace: 'cypress', aliases: ['test:open'])]
function open(): void
{
    if (!e2e_dir_exists()) {
        io()->note('No application/tests/E2E directory, Cypress open skipped.');

        return;
    }

    io()->title('Opening Cypress');
    docker_compose(['run', '--rm', 'cypress-open'], profiles: ['default', 'test']);
}
