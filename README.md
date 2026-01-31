# Better Actions

> This project is still in very very early stages of development.

## Yet Another CI/CD?

Better Actions is a project aimed to open up and extend the standard created by GitHub Actions. A lot of time and effort went into GitHub Action from the community, but there's a sentiment online that Microsoft kind of neglected the runtime on which everything is based on.

The idea behind this project is to allow advanced usage of the
vast amount of available github actions, workflows and compatible
scripts without needing to constrain ourselves to the restrictions posed by the runtime.

It is intended to run using your existing GitHub runners.

## Why is GitHub Actions not enough?

- Better controls over concurrency
  - The current concurrency available in GHA is only on the Job level. I want to add support on the step level.
- Jobs should be able to start running before another Job completely finishes its work.
  - Github chooses a pessimistic approach to concurrency control - if a Job B `needs` another Job A, Job A must fully finish running before Job B can start.
  - This is useful for example if Job B builds multiple artifacts sequentially, and Job A is a matrix that tests all these artifacts.
  - Another option is just to start "warming" the machine of Job B with a got clone, npm install etc.
- Fine Grained Cancellation Controls
  - Github offers just one button - "Cancel Workflow" - which cancels all jobs in the workflow. This is not always desirable.
  - Sometimes jobs are flakey, and you want to rerun a failed job while the other jobs in the pipeline are still running. Github Doesn't offer a way to do this.
-
