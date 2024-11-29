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
