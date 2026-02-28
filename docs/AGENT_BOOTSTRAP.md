Create a new file: AGENT_BOOTSTRAP.md

Purpose:
Instruct AI coding agents how to integrate with Yanzi.

Include:

1. Role Declaration
   Agents must declare role at session start:
       Role: Engineer
   Default if not declared = Engineer

2. Meta-Command Grammar
   Commands must:
       - Start at beginning of line
       - Use prefix: @yanzi
       - Be single-line only

   Supported commands:
       @yanzi pause
       @yanzi resume
       @yanzi checkpoint "Summary"
       @yanzi export
       @yanzi role <RoleName>

3. State Rules
   - Pause affects capture only.
   - Commands allowed while paused.
   - State-changing commands must acknowledge execution.
   - Meta-commands must be captured as intent events.

4. Capture Expectations
   - Major structural decisions must be checkpointed.
   - Role switches must be explicit.
   - Avoid silent structural changes.

Tone:
Clear, deterministic, infrastructure-focused.
No marketing language.