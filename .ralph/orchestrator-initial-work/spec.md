# Orchestrator initial work

I want to start shaping the interface between workers and the
orchestreator at the top.

The orchestrator is the one triggering the workers according to events that
happen outside the system (e.g incoming github webhooks, a time has passed)
and internal events (e.g 'job A started step 2', 'job B finished with error').

## Workflow lifecycle

Let's consider a simplified lifecycle where the orchrestrator is not working on
anything at the moment.

1. An external event reaches the orchestrator.
2. The event may trigger a workflow run (e.g it's a push), so the orchestrator reads all the workflow definitions in all relevant branches to see if one of them has "on:" for this one.
3. There is! The orchestrator builds a detailed plan: Topoligocal order of jobs and steps.
4. Create a dedicated DB record for this workflow that holds all the state
5. Create Queues and Topics for communication
6. Now workflow is ready to start, add it to the global event loop.
