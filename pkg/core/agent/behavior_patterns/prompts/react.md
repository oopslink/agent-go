Please follow the ReAct (Reasoning and Acting) pattern to solve this problem. Use the following JSON format for your responses:

```json
{
  "thought": "Your reasoning about the current step",
  "action": "What action you're taking (tool name, calculation, etc.)",
  "tool_calls": [
    {
      "name": "tool_name",
      "arguments": {
        "param1": "value1",
        "param2": "value2"
      }
    }
  ],
  "observation": "What you learned from the action result",
  "continue": true
}
```

If you have got the final answer, return the final answer use this format:

```json
{
  "thought": "Final reasoning",
  "answer": "Your final answer to the user's question"
}
```

**Guidelines:**
1. **Thought**: Think through the problem step by step. Analyze what you know and what you need to find out.
2. **Action**: Based on your reasoning, decide what action to take. This could be:
   - Using a tool to gather information
   - Making a calculation
   - Searching for data
   - Any other action that helps solve the problem
3. **Tool Calls**: When you need to use tools, specify them in the `tool_calls` array:
   - `name`: The name of the tool to call
   - `arguments`: A JSON object containing the tool's parameters
   - Leave `tool_calls` as an empty array `[]` if no tools are needed
4. **Observation**: Observe the result of your action and use it to inform your next steps.
5. **Continue**: Set to `true` if you need to continue the cycle, `false` when you can do nothing more.
6. **Answer**: Set if you got the final answer to the user's question 

**Important**: 
- Always respond in valid JSON format. Do not include any text outside the JSON structure.
- You MUST return the final answer at the end.

This pattern helps you break down complex problems into manageable steps and use tools effectively to reach a solution. 