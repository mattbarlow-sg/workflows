# Instructions
Always stop and ask the user before implementing a workaround.

Examples:
WRONG: "Package XYZ is not installed, therefore I must add a warning and ignore Package XYZ if it does not exist."
RIGHT: "Package XYZ is not installed, therefore I must ask the user if they want to install it."

WRONG: "We changed our types and can no longer ingest the persisted data, therefore I must update the program to support both types."
RIGHT: "We changed our types and can no longer ingest the persisted data, therefore I must ask the user if they want to delete the old data."

## MPC (Model Predictive Control) Philosophy

In robotics, Model Predictive Control (MPC) plans a full trajectory, executes a single control input, then replans from the new state—a rolling-horizon strategy that keeps the robot agile in uncertain environments. This approach brings the same principle to software planning: every node carries a 0-to-1 materialization score expressing how "locked-in" it is.

### Materialization Score Progression

The materialization score quantifies how concrete and committed a plan node is:

- **0.1**: Initial planning - node identified but highly flexible
- **0.2-0.3**: BPMN diagrams added - workflow structure defined
- **0.4-0.5**: Specifications written - requirements locked in
- **0.6-0.7**: Tests created - behavior contracts established  
- **0.8-0.9**: Implementation complete - code written and tested
- **1.0**: Fully materialized - deployed and validated in production

As you add artifacts (BPMN, specs, tests, code), the materialization score increases, reflecting decreased flexibility but increased confidence. The entry node typically has the highest score (ready to execute), while downstream nodes decay toward lower scores, signaling openness to change.

This quantifies the insight from "Why Greatness Cannot Be Planned": rigid end-to-end blueprints invite failure, whereas flexible, step-wise commitments let emergent opportunities guide the path to success. The system stays agile by keeping distant future nodes at low materialization, allowing pivots based on lessons learned during execution.
