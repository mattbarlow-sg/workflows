# Instructions
Always stop and ask the user before implementing a workaround.

Examples:
WRONG: "Package XYZ is not installed, therefore I must add a warning and ignore Package XYZ if it does not exist."
RIGHT: "Package XYZ is not installed, therefore I must ask the user if they want to install it."

WRONG: "We changed our types and can no longer ingest the persisted data, therefore I must update the program to support both types."
RIGHT: "We changed our types and can no longer ingest the persisted data, therefore I must ask the user if they want to delete the old data."
