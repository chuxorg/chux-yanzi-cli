# CAP-200: Operational Supervision for Autonomous and Semi-Autonomous Work

## Problem Statement

Yanzi is moving toward a world in which useful work is increasingly performed by a mix of humans, AI agents, automation systems, and other execution runtimes.

As autonomous and semi-autonomous systems become capable of performing useful work, the scarce capability becomes supervision: knowing what is happening, why it is happening, whether it should continue, and how to recover or intervene.

This problem is visible first in software engineering because software work already produces artifacts, logs, commands, reviews, and operational traces that can be inspected. That makes it a strong credibility domain. But the underlying problem is broader than software engineering. Any domain with delegated work, partial automation, asynchronous execution, or shared responsibility eventually requires the same supervisory capabilities.

The problem is not how to make systems act on their own. The problem is how operators remain able to observe, understand, govern, and recover delegated work once execution is distributed across multiple participants and runtimes.

Yanzi should be understood in that context. It records, observes, preserves, and supports supervision. It does not decide what should happen next.

## What Changed

Historically, many tools assumed that work was either directly performed by a human in a visible interface or executed by a narrow automation system with a bounded purpose. That assumption is weakening.

Useful work can now be:

- delegated to another human asynchronously
- executed partially by scripts, pipelines, or service runtimes
- advanced by LLM-backed tools or agentic systems
- resumed across sessions, machines, and contexts
- passed between people and machines without a single continuous operator present

This changes the supervisory burden.

The core difficulty is no longer only task creation or task execution. The difficulty is maintaining reliable visibility and governance while work moves across actors, tools, environments, and time.

Operators now need to know not only that work was assigned, but also:

- what the worker believed it was doing
- what evidence it used
- what changed over time
- where the current state actually stands
- what can be trusted after interruption, failure, or handoff

That is a materially different operational requirement than traditional task tracking, logging, or workflow status.

## Why Existing Tools Are Insufficient

Existing tools usually capture only one slice of the problem.

Issue trackers capture intent and assignment, but usually not live execution state, evidence trails, recovery context, or machine-participant handoffs.

Workflow engines capture ordered steps and automation logic, but usually assume that the system already knows what should happen next and that execution can be reduced to predefined transitions.

Observability stacks capture telemetry, logs, traces, and metrics, but usually do not preserve operator intent, approvals, decisions, accountability boundaries, or resumable work context.

Chat transcripts capture conversation, but usually do not preserve structured operational state well enough to support interruption, recovery, verification, or governance.

Agent frameworks capture prompts, tools, and execution loops, but usually center the agent runtime itself rather than the broader supervisory problem across humans and machines.

Traditional project systems capture planning, ownership, and status reporting, but usually lag behind work in progress and do not provide a trustworthy operational record of what actually happened.

The gap is not the absence of execution systems. The gap is the absence of a durable supervisory surface spanning mixed participants, mixed runtimes, and mixed timescales.

## Definition of Supervision

For this CAP, supervision means the operational ability to govern delegated work without becoming the executor of that work.

Supervision is the ability to:

- observe work in progress
- understand current operational state
- inspect intent, evidence, and decisions
- interrupt or pause execution
- resume from a known state
- recover after failure or context loss
- verify what happened
- preserve an audit trail
- assign accountability
- support human governance

Supervision is not mere visibility. A scrolling log is not sufficient. A conversation transcript is not sufficient. A status badge is not sufficient.

Supervision requires that the work can be inspected as an operational record with enough preserved context to support intervention, review, recovery, and accountability.

Supervision also does not imply autonomous control. The supervisor may choose to continue, pause, redirect, approve, reject, or hand off. Yanzi's role is to support those acts by preserving the state and evidence required to make them possible.

## Definition of Operational State

For supervised work to be governable, it must have a legible operational state.

Operational state includes:

- current work item
- actor or worker
- intent
- inputs
- outputs
- evidence
- status
- timeline
- decisions
- approvals
- errors
- cost/usage
- checkpoints
- recovery context

This definition matters because many systems expose only fragments of state. A tool may show outputs without intent, or status without evidence, or errors without recovery context.

Operational state is the minimum durable record required for an operator to answer what is happening now, how the system arrived here, and what can safely happen next.

It must be preserved across interruption, reassignment, failure, and context loss. If the state cannot survive those moments, the work is not meaningfully supervisable.

## Actors and Responsibilities

Supervised work involves multiple actors with distinct responsibilities.

- operators are responsible for oversight, intervention, governance, and final accountability
- workers are responsible for performing delegated work; workers may be humans, AI agents, scripts, services, pipelines, or other execution runtimes
- requesters define the work to be done, the desired outcome, and the constraints under which it should proceed
- reviewers or approvers validate outputs, decisions, or transitions when formal governance is required
- platform or system owners define the boundaries, retention rules, access controls, and operational policies that make supervision possible

These roles may overlap in practice. One person may be both requester and operator. One system may both execute and emit evidence. But the responsibilities should remain conceptually distinct.

Yanzi should preserve the relationship between actor, action, evidence, and authority. It should not collapse all participants into a generic execution stream.

## Lifecycle of Supervised Work

Supervised work typically moves through a recurring lifecycle rather than a single linear run.

1. Work is defined with an intent, scope, and responsible actors.
2. Execution begins under some runtime, participant, or delegated worker.
3. State, evidence, and outputs accumulate while the work is in progress.
4. An operator or reviewer inspects progress and determines whether execution should continue as-is.
5. Work may be paused, interrupted, redirected, handed off, or approved.
6. Failures, ambiguities, or context loss may require recovery from a known checkpoint or preserved state.
7. Execution resumes, completes, or terminates.
8. The final record must support verification, audit, accountability, and retrospective understanding.

This lifecycle is important because supervised work is rarely a clean request-response transaction.

It spans time. It crosses tools. It may switch workers. It may fail partially. It may require human intervention after machine progress, or machine continuation after human review.

Any system meant to support supervision must treat these transitions as first-class operational events rather than edge cases.

## Core Questions an Operator Must Answer

An operator supervising delegated work must be able to answer a small set of operational questions quickly and reliably:

- what work is currently in progress
- who or what is performing it
- what is the stated intent
- what inputs, tools, and evidence are being used
- what outputs have been produced so far
- what decisions or approvals have occurred
- what changed most recently
- whether the work is healthy, blocked, failing, or ambiguous
- whether execution should continue, pause, or be interrupted
- how to resume safely from the current point
- how to recover if context or execution is lost
- what exactly happened after the fact
- who is accountable for the result and for each decision boundary

If those questions cannot be answered without reconstructing history from scattered tools, supervision is weak even if execution appears productive.

## Non-Goals

This CAP does not propose that Yanzi become any of the following:

- Yanzi is not an AI model
- Yanzi is not an agent framework
- Yanzi is not a workflow engine
- Yanzi is not an IDE
- Yanzi is not a SaaS control plane
- Yanzi is not responsible for deciding what should happen next

This CAP also does not define an implementation architecture, product packaging model, hosted operating model, or runtime orchestration strategy.

It does not assume that every workflow should be automated, that every participant should be agentic, or that every decision can be encoded into a deterministic state machine.

The purpose here is narrower and more important: define the problem space clearly enough that future product decisions can be evaluated against the supervisory need instead of drifting toward execution-centric abstractions.

## Initial Implications for Yanzi

If Yanzi is to support operational supervision well, several implications follow.

First, Yanzi should treat operational records as a primary artifact, not as incidental exhaust from execution.

Second, Yanzi should preserve state in a way that survives interruption, handoff, review, and recovery. The preserved record must be useful both during execution and after the fact.

Third, Yanzi should model work in a participant-neutral way. Humans, agents, scripts, and services should all be representable as execution participants without making any one class of worker the center of the system.

Fourth, Yanzi should privilege legibility over autonomy. If a capability makes execution faster but makes supervision weaker, the tradeoff should be explicit and usually rejected.

Fifth, Yanzi should support intervention boundaries without claiming decision authority. Pause, resume, inspect, verify, and recover are aligned with Yanzi's role. Deciding what should happen next is not.

Sixth, Yanzi should preserve accountability and governance context alongside work state. The operational record is incomplete if it cannot answer who acted, under what authority, with what evidence, and with what result.

Finally, Yanzi should continue to build credibility in software engineering while keeping the conceptual model general. Software is the first proving ground, not the definition of the category.
