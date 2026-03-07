# Pipeline Form POC

A demonstration of form-based pipeline workflows in Haira, showcasing multi-step execution, progress tracking, and rich UI components.

## Features

- **Form-based Configuration**: Web UI forms for pipeline setup
- **Multi-step Execution**: Step-by-step pipeline execution with progress tracking
- **Rich UI Components**: Progress bars, status cards, key-value displays, tables
- **Error Handling**: Graceful error handling with user-friendly messages
- **Interactive Chat**: AI assistant for pipeline help and troubleshooting
- **File Upload**: Support for file uploads and processing
- **Validation**: Input validation with detailed error messages

## Workflows

### 1. Pipeline Form (`/api/pipeline`)
Main pipeline configuration and execution workflow:
- **Mode**: Form
- **Features**: Full pipeline configuration with validation, multi-step execution, progress tracking
- **Parameters**:
  - `name`: Pipeline name (required)
  - `description`: Pipeline description
  - `priority`: Execution priority (low/medium/high)
  - `environment`: Target environment (dev/staging/prod)
  - `notify_on_completion`: Send notifications when done
  - `max_retries`: Maximum retry attempts
  - `input_file`: Optional file upload

### 2. Quick Pipeline (`/api/quick`)
Simplified form for quick tasks:
- **Mode**: Form
- **Features**: Minimal configuration for simple tasks
- **Parameters**:
  - `task_name`: Name of the task to execute
  - `data_source`: Data source (sample/upload/database)
  - `output_format`: Output format (csv/json/xml)

### 3. Chat Assistant (`/api/chat`)
Interactive AI assistant:
- **Mode**: Chat
- **Features**: Help with configuration, troubleshooting, status queries
- **Tools**: Configuration validation, step execution, report generation

## UI Components Demonstrated

### Forms
```haira
@webui(
    title: "Pipeline Form",
    description: "Configure and execute data processing pipelines",
    mode: "form"
)
workflow PipelineForm(
    name: string,
    description: string,
    priority: string = "medium",
    // ... more parameters
) -> { /* results */ }
```

### Progress Tracking
```haira
log.render(ui.progress("Pipeline Execution", [
    {"name": "Validate Configuration", "status": "done"},
    {"name": "Initialize Pipeline", "status": "running"},
    {"name": "Process Data", "status": "pending"},
    {"name": "Generate Output", "status": "pending"},
    {"name": "Cleanup", "status": "pending"}
]))
```

### Status Cards
```haira
log.render(ui.status_card("success", "Pipeline Completed", "All steps executed successfully"))
log.render(ui.status_card("error", "Validation Failed", "Pipeline name is required"))
log.render(ui.status_card("warning", "Partial Success", "Some steps completed with warnings"))
```

### Key-Value Displays
```haira
log.render(ui.key_value("Pipeline Configuration", {
    "Name": name,
    "Priority": priority,
    "Environment": environment,
    "Max Retries": conv.to_string(max_retries)
}))
```

### Grouped Components
```haira
log.render(ui.group(
    ui.status_card("success", "Pipeline Completed"),
    ui.key_value("Results", { /* ... */ }),
    ui.progress("Steps", [ /* ... */ ])
))
```

## Pipeline Steps

1. **Validate Configuration**: Check input parameters and business rules
2. **Initialize Pipeline**: Set up execution environment and resources
3. **Process Data**: Execute main data processing logic
4. **Generate Output**: Create output files and reports
5. **Cleanup**: Clean up temporary resources and finalize

## Error Handling

The POC demonstrates several error handling patterns:

- **Input Validation**: Check required fields and valid values
- **Step Failures**: Handle individual step failures with retry logic
- **Global Error Handler**: `onerror` blocks for unexpected failures
- **User-Friendly Messages**: Clear error messages with actionable guidance

## Running the POC

1. **Start the server**:
   ```bash
   haira run poc/pipeline-form/main.haira
   ```

2. **Access the UIs**:
   - Pipeline Form: http://localhost:9005/_ui/api/pipeline
   - Quick Pipeline: http://localhost:9005/_ui/api/quick
   - Chat Assistant: http://localhost:9005/_ui/api/chat
   - Health Check: http://localhost:9005/api/health
   - Dashboard: http://localhost:9005/_observe

3. **Try the workflows**:
   - Fill out the pipeline form with different configurations
   - Upload files to test file processing
   - Use the chat assistant to ask questions
   - Try invalid inputs to see validation in action

## Environment Variables

Optional environment variables:
- `OPENAI_API_KEY`: For the chat assistant (uses gpt-4o-mini)

## Architecture Patterns

### Form Workflows
- Use `mode: "form"` in `@webui` annotation
- Define parameters with types and defaults
- Return structured results
- Use `log.render()` for UI updates during execution

### Progress Tracking
- Use `ui.progress()` to show step-by-step progress
- Update progress state as steps complete
- Show different statuses: pending, running, done, failed

### Validation
- Validate inputs early in the workflow
- Use tools for complex validation logic
- Return clear error messages with specific guidance
- Use status cards to communicate validation results

### Error Recovery
- Use `onerror` blocks for graceful error handling
- Provide retry mechanisms where appropriate
- Log errors for debugging while showing user-friendly messages
- Allow partial completion where possible

## Extending the POC

This POC can be extended with:

- **Real Data Processing**: Replace simulated steps with actual data processing
- **Database Integration**: Add database connections for data persistence
- **External APIs**: Integrate with external services and APIs
- **Advanced Validation**: Add more sophisticated validation rules
- **Scheduling**: Add pipeline scheduling capabilities
- **Monitoring**: Add real-time monitoring and alerting
- **Templates**: Add pipeline templates for common use cases