# ECS Task Runner

## Example

Add the following lines to your `pipeline.yml`:

```yml
steps:
  - plugins:
      - cultureamp/migrations-runner#v0.0.0:
          parameter-name: "test-parameter"
          command: "/bin/migrate"
          timeout: 900
```

## Configuration

### `parameter-name` (Required, string)
The name or ARN of the parameter in Parameter Store that contains the task definition.

### `command` (Optional, string)
The name of the command to run in the task. When omitted, the task will run the command specified in the parameter.

### `timeout` (Optional, integer)
The timeout in seconds that the plugin will wait for the task to complete. If the task does not complete within this time, the plugin will fail. The task execution will continue to run in the background.

Default: 2700

## Usage
This plugin is based on an existing pattern in `murmur` where database migrations are run as a task on ECS. To provide additional context for how this plugin is expected to be used, this is the expected pattern:

- A CI image is built and pushed to ECR
- The entrypoint of the image is overridden in the task definition to run the specific migration task

# Requisite Infrastructure

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
