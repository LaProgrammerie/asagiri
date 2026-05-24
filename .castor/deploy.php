<?php

namespace deploy;

use Castor\Attribute\AsTask;

use function Castor\context;
use function Castor\io;
use function Castor\run;
use function Castor\variable;

function yoimachi_dir(): string
{
    return variable('root_dir') . '/infrastructure/yoimachi';
}

function ensure_yoimachi_binary(): void
{
    $c = context()->withAllowFailure()->withQuiet();
    $result = run(['bash', '-lc', 'command -v yoimachi'], context: $c);
    if (!$result->isSuccessful()) {
        throw new \RuntimeException('yoimachi CLI not found. Install it first (see infrastructure/yoimachi/README.md).');
    }
}

function ensure_yoimachi_config_exists(): void
{
    $dir = yoimachi_dir();
    if (!is_dir($dir)) {
        throw new \RuntimeException('Missing infrastructure/yoimachi directory.');
    }

    $yaml = $dir . '/yoimachi.yaml';
    $example = $dir . '/yoimachi.yaml.example';

    if (is_file($yaml)) {
        return;
    }

    if (!is_file($example)) {
        throw new \RuntimeException('Missing yoimachi.yaml and yoimachi.yaml.example.');
    }

    copy($example, $yaml);
    io()->note('Created infrastructure/yoimachi/yoimachi.yaml from example. Adapt values for your project.');
}

#[AsTask(description: 'Validate Yoimachi config', aliases: ['yoi:validate'])]
function validate(): void
{
    ensure_yoimachi_binary();
    ensure_yoimachi_config_exists();

    io()->title('Yoimachi validate');
    run(['yoimachi', 'validate'], context: context()->withWorkingDirectory(yoimachi_dir()));
}

#[AsTask(description: 'Initialize deployment workflow (Yoimachi)', aliases: ['init'])]
function init(): void
{
    validate();
}

#[AsTask(description: 'Generate deployable infra from Yoimachi', aliases: ['yoi:generate'])]
function generate(): void
{
    ensure_yoimachi_binary();
    ensure_yoimachi_config_exists();

    io()->title('Yoimachi generate');
    run(['yoimachi', 'generate'], context: context()->withWorkingDirectory(yoimachi_dir()));
}

#[AsTask(description: 'Plan deployment (Yoimachi generate)', aliases: ['plan'])]
function plan(): void
{
    generate();
}

#[AsTask(description: 'Apply deployment if supported by Yoimachi CLI', aliases: ['apply'])]
function apply(): void
{
    ensure_yoimachi_binary();
    ensure_yoimachi_config_exists();

    io()->title('Yoimachi apply');
    $apply = run(
        ['bash', '-lc', 'yoimachi up'],
        context: context()->withWorkingDirectory(yoimachi_dir())->withAllowFailure()
    );
    if (!$apply->isSuccessful()) {
        io()->note('`yoimachi up` not available or failed on this version. Validation + generation have already been executed.');
    }
}

#[AsTask(description: 'Validate then generate Yoimachi infra', aliases: ['yoi:deploy', 'deploy'])]
function deploy(): void
{
    init();
    plan();
    apply();
    io()->success('Yoimachi deployment artifacts are ready.');
}
