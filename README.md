# Migrations Runner Plugin

This plugin is designed to run database migrations tasks, without requiring consumers to write and maintain their own scripts including execution and tracking compute services to facilitate migration.

> [!NOTE]
> Note that this plugin only executes migration tasks, and assumes that the required infrastructure has been deployed. For specific information on how to configure and deploy the prerequisite infrastructre, please refer to the steps outlined here to setup and configure the `MigrationsRunner` construct in `cdk-constructs`.

## Requisite Infrastructure

This plugin comes with some assumed infrastructure that needs to be deployed before it can be run. This infrastructure is as follows:

- An ECS cluster
- An ECS task definition
- An ECR image
- An IAM role for the ECS task
- An IAM role for the BK agent to start the task
- A Parameter Store parameter extending the task definition by providing entrypoint overrides and networking configuration
- A log group for the task
- A security group for your service (this can be the [base-infrastructure-for-services](https://github.com/cultureamp/base-infrastructure-for-services) source security group)

This can be visualised below:

![The overall flow of this plugin and AWS resources](docs/images/diagram.svg)

## Example

Add the following lines to your `pipeline.yml`:

```yml
steps:
  - label: "Run my very cool migration task"
    plugins:
      - cultureamp/aws-assume-role:
        role: arn:aws:iam::${AWS_ACCOUNT_ID}:role/deploy-role-cool-service
        region: ${REGION}
      - cultureamp/migrations-runner#v1.0.0:
          parameter-name: "/cool-service/cool-farm/migrations-runner-config"
          command: "/bin/migrate"
```

## Configuration

### `parameter-name` (Required, string)

The name of the parameter in Parameter Store that contains the task definition. This will be setup by the `MigrationsRunner` construct, so refer to the stack where you use `MigrationsRunner` to find your specific parameter name. The parameter created by the construct will always end in `/migrations-runner-config`.

### `command` (Optional, string)

The name of the command or script to run in the task. When omitted, the task will run the command specified in the container's `CMD` or `ENTRYPOINT`.

> [!NOTE]
> Care must be taken, as depending on whether the Dockerfile used to create the referenced image has a `CMD` or `ENTRYPOINT` instruction, the command to execute in the container will either be overridden or appended to, respectively.
>
> To illustrate these behaviours, consider the following scenarios:
>
> - A Dockerfile with an `ENTRYPOINT` of `./run-binary`. Leaving the plugin's command argument blank will simply execute `./run-binary` in the container. Inputting `run-sub-command` in the plugin's command argument will execute `./run-binary run-sub-command`.
> - A Dockerfile with a `CMD` of `./run-binary`. Leaving the plugin's command argument blank will simply execute `./run-binary` in the container. Inputting `run-sub-command` in the plugin's command argument will execute `run-sub-command`.

> [!TIP]
> When specifying a command to run, ensure that the command is:
>
> - Enclosed in quotes
> - Space-separated
>
> For example if you'd like to run the following command for your migration task: `parameter-store-exec bundle exec bin/run_mongodb_migrations`
>
> You'd configure the plugin like this:
>
>```yml
>steps:
>  - plugins:
>      - cultureamp/migrations-runner#v1.0.0:
>          parameter-name: "/cool-service/cool-farm/migrations-runner-config"
>          command: "parameter-store-exec bundle exec bin/run_mongodb_migrations"
>```

### `timeout` (Optional, integer)

The timeout in seconds that the plugin will wait for the task to complete. If the task does not complete within this time, the plugin will fail. The task execution will continue to run in the background.

> [!NOTE]
> This only affects how the migration task's status is reported in the Buildkite step, and the migration task may still run successfully even if the plugin timeouts. To avoid false negatives on migration task failures, we recommend using the default value.

Default: 2700

## Context

This plugin is based on an existing pattern in `murmur` where database migrations are run as a task on ECS. To provide additional context for how this plugin is expected to be used, this is the expected pattern:

- A CI image is built and pushed to ECR
- The command of the image can be overridden in the task definition to run the specific migration task

This plugin was developed to remove the requirement for engineers to write their own code to perform database migrations, and to remove any variations between these methods.
