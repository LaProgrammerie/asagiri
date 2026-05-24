<?php

namespace qa;

use Castor\Attribute\AsOption;
use Castor\Attribute\AsRawTokens;
use Castor\Attribute\AsTask;

use function Castor\io;
use function Castor\variable;
use function docker\docker_compose_run;
use function docker\docker_exit_code;

function has_tool(string $path): bool
{
    return is_file(variable('root_dir') . '/' . $path);
}

#[AsTask(description: 'Runs all QA tasks')]
function all(): int
{
    return max(
        qa_js(),
        static_checks(),
        phpunit(),
    );
}

#[AsTask(description: 'Installs tooling')]
function install(): void
{
    io()->title('Installing QA tooling');

    docker_compose_run('composer install -o', workDir: '/var/www/tools/php-cs-fixer');
    docker_compose_run('composer install -o', workDir: '/var/www/tools/phpstan');
    docker_compose_run('composer install -o', workDir: '/var/www/tools/twig-cs-fixer');
    if (has_tool('tools/rector/composer.json')) {
        docker_compose_run('composer install -o', workDir: '/var/www/tools/rector');
    }
    if (has_tool('tools/deptrac/composer.json')) {
        docker_compose_run('composer install -o', workDir: '/var/www/tools/deptrac');
    }
    if (has_tool('tools/phparkitect/composer.json')) {
        docker_compose_run('composer install -o', workDir: '/var/www/tools/phparkitect');
    }
}

#[AsTask(description: 'Updates tooling')]
function update(): void
{
    io()->title('Updating QA tooling');

    docker_compose_run('composer update -o', workDir: '/var/www/tools/php-cs-fixer');
    docker_compose_run('composer update -o', workDir: '/var/www/tools/phpstan');
    docker_compose_run('composer update -o', workDir: '/var/www/tools/twig-cs-fixer');
    if (has_tool('tools/rector/composer.json')) {
        docker_compose_run('composer update -o', workDir: '/var/www/tools/rector');
    }
    if (has_tool('tools/deptrac/composer.json')) {
        docker_compose_run('composer update -o', workDir: '/var/www/tools/deptrac');
    }
    if (has_tool('tools/phparkitect/composer.json')) {
        docker_compose_run('composer update -o', workDir: '/var/www/tools/phparkitect');
    }
}

#[AsTask(description: 'Runs static checks (phpstan/cs/twig-cs + optional deptrac/phparkitect/rector)')]
function static_checks(): int
{
    $codes = [
        phpstan(),
        cs(true),
        twigCs(true),
    ];
    if (has_tool('tools/deptrac/composer.json')) {
        $codes[] = deptrac();
    }
    if (has_tool('tools/phparkitect/composer.json')) {
        $codes[] = phparkitect();
    }
    if (has_tool('tools/rector/composer.json')) {
        $codes[] = rector(true);
    }

    return max(...$codes);
}

/**
 * @param string[] $rawTokens
 */
#[AsTask(description: 'Runs PHPUnit when available', aliases: ['phpunit'])]
function phpunit(#[AsRawTokens] array $rawTokens = []): int
{
    if (!is_file(variable('root_dir') . '/application/vendor/bin/phpunit')) {
        io()->note('PHPUnit not installed yet (application/vendor/bin/phpunit missing), skipping.');

        return 0;
    }

    io()->section('Running PHPUnit...');

    return docker_exit_code('vendor/bin/phpunit ' . implode(' ', $rawTokens), workDir: '/var/www/application');
}

#[AsTask(description: 'Runs PHPStan', aliases: ['phpstan'])]
function phpstan(
    #[AsOption(description: 'Generate baseline file', shortcut: 'b')]
    bool $baseline = false,
): int {
    if (!is_dir(variable('root_dir') . '/tools/phpstan/vendor')) {
        install();
    }

    io()->section('Running PHPStan...');

    $options = $baseline ? '--generate-baseline --allow-empty-baseline' : '';
    $command = \sprintf('phpstan analyse --memory-limit=-1 %s -v', $options);

    return docker_exit_code($command, workDir: '/var/www');
}

#[AsTask(description: 'Fixes Coding Style', aliases: ['cs'])]
function cs(bool $dryRun = false): int
{
    if (!is_dir(variable('root_dir') . '/tools/php-cs-fixer/vendor')) {
        install();
    }

    io()->section('Running PHP CS Fixer...');

    if ($dryRun) {
        return docker_exit_code('php-cs-fixer fix --dry-run --diff', workDir: '/var/www');
    }

    return docker_exit_code('php-cs-fixer fix', workDir: '/var/www');
}

#[AsTask(description: 'Fixes Twig Coding Style', aliases: ['twig-cs'])]
function twigCs(bool $dryRun = false): int
{
    if (!is_dir(variable('root_dir') . '/tools/twig-cs-fixer/vendor')) {
        install();
    }

    io()->section('Running Twig CS Fixer...');

    if ($dryRun) {
        return docker_exit_code('twig-cs-fixer', workDir: '/var/www');
    }

    return docker_exit_code('twig-cs-fixer --fix', workDir: '/var/www');
}

#[AsTask(description: 'Runs Deptrac when configured', aliases: ['deptrac'])]
function deptrac(): int
{
    if (!has_tool('tools/deptrac/composer.json')) {
        io()->note('Deptrac not configured in tools/deptrac, skipping.');

        return 0;
    }
    if (!is_dir(variable('root_dir') . '/tools/deptrac/vendor')) {
        install();
    }

    return docker_exit_code('php tools/deptrac/vendor/bin/deptrac analyse --no-progress', workDir: '/var/www');
}

#[AsTask(description: 'Runs PHPArkitect when configured', aliases: ['phparkitect'])]
function phparkitect(): int
{
    if (!has_tool('tools/phparkitect/composer.json')) {
        io()->note('PHPArkitect not configured in tools/phparkitect, skipping.');

        return 0;
    }
    if (!is_dir(variable('root_dir') . '/tools/phparkitect/vendor')) {
        install();
    }

    return docker_exit_code('php tools/phparkitect/vendor/bin/phparkitect check', workDir: '/var/www');
}

#[AsTask(description: 'Runs Rector when configured', aliases: ['rector'])]
function rector(bool $dryRun = true): int
{
    if (!has_tool('tools/rector/composer.json')) {
        io()->note('Rector not configured in tools/rector, skipping.');

        return 0;
    }
    if (!is_dir(variable('root_dir') . '/tools/rector/vendor')) {
        install();
    }

    $mode = $dryRun ? 'process --dry-run' : 'process';

    return docker_exit_code('php ../tools/rector/vendor/bin/rector ' . $mode . ' --no-progress-bar', workDir: '/var/www/application');
}

#[AsTask(description: 'Runs JS QA if package.json exists', aliases: ['qa-js'])]
function qa_js(): int
{
    if (!is_file(variable('root_dir') . '/application/package.json')) {
        io()->note('No application/package.json, JS QA skipped.');

        return 0;
    }

    return docker_exit_code('npm run qa:js', workDir: '/var/www/application');
}
