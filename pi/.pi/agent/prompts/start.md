First, use the `orchestrator` to analyze my request. Its job is to break it down into steps, determine if we have the right agents, ask me clarifying questions to refine the task, and finally propose a structured delegation plan (which agents to run and in what order). 

If `orchestrator` determines we are missing a specific agent (e.g. for a new language like Java), it should STOP and tell me to run `@meta` to create the missing agent.

My request: {query}
