# ECS Task Runner

## Example

Add the following lines to your `pipeline.yml`:

```yml
steps:
  - plugins:
      - cultureamp/ecs-task-runner#v0.0.0:
          message: "This is the message that will be annotated!"
```

## Configuration

### `message` (Required, string)

The message to annotate onto the build.


# Requisite Infrastructure

This plugin comes with some assumed infrastructure that needs to be deployed before it can be run. This infrastructure is as follows:

- An ECS cluster
- An ECS task definition
- An ECR image
- An IAM role for the ECS task
- An IAM role for the BK agent to start the task
- A Parameter Store parameter defining the task configuration (including security groups, etc)
- A log group for the task

This can be visualised below:
![The overall flow of this plugin and AWS resources](docs/images/diagram.png)
