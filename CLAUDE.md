# Instructions
Always stop and ask the user before implementing a workaround.

Examples:

WRONG: "Package XYZ is not installed, therefore I must add a warning and ignore Package XYZ if it does not exist."
RIGHT: "Package XYZ is not installed, therefore I must ask the user if they want to install it."

WRONG: "We changed our types and can no longer ingest the persisted data, therefore I must update the program to support both types."
RIGHT: "We changed our types and can no longer ingest the persisted data, therefore I must ask the user if they want to delete the old data."

## MPC (Model Predictive Control) Philosophy

In robotics, Model Predictive Control (MPC) plans a full trajectory, executes a single control input, then replans from the new state a rolling-horizon strategy that keeps the robot agile in uncertain environments. This approach brings the same principle to software planning: every node carries a 0-to-1 materialization score expressing our confidence in executing that step.

### Materialization Score: Confidence in Execution

The materialization score quantifies **how confident we are that a particular step in the plan will be executed as currently envisioned**:

- **1.0**: Complete confidence - we know we can execute this step exactly as planned (like seeing the next stepping stone clearly through the fog)
- **0.5**: Moderate confidence - we think we can execute this step, but there are known uncertainties (engine might not start, path might be blocked)
- **0.1**: Low confidence - this step is highly speculative and will likely change based on what we learn from earlier steps (distant stones we can't yet see clearly)

The materialization score is NOT about implementation completeness or how much code is written. It's about **planning confidence and uncertainty**.

### Example: Software Development Plan

1. **Set up project structure** (materialization: 1.0) - We're certain we can do this next
2. **Create user authentication** (materialization: 0.7) - Fairly confident, but depends on framework choice from step 1
3. **Add payment processing** (materialization: 0.4) - Less certain, depends on authentication design and business requirements we'll discover
4. **Scale to 10K users** (materialization: 0.2) - Highly speculative, will depend on architecture choices and learnings from previous steps

The immediate next node typically has the highest score (ready to execute), while downstream nodes have lower scores, reflecting our uncertainty about distant future steps.

This quantifies the insight from "Why Greatness Cannot Be Planned": rigid end-to-end blueprints invite failure, whereas flexible, step-wise commitments let emergent opportunities guide the path to success. The system stays agile by keeping distant future nodes at low materialization, allowing pivots based on lessons learned during execution.
