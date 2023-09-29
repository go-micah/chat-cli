# chat-cli

A little terminal based program that talks to the Amazon Bedrock API.

## Prerequisites

1. You will need an AWS account.
2. You will need to enable Cloud v2 in Amazon Bedrock
3. You will need to run `aws config` from the command line to set up your default profile

## Compile

    $ make cli

## Run

    $ ./bin/chat-cli

## Use

You can send messages to Claude in natural language. If the response is lengthy it may take a little while to appear. The reposnes are currently capped at 300 tokens, so things may get truncated if Claude gets long winded.

Type `quit` or `ctl-c` to quit.