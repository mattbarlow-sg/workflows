---
Description: Capture a lesson for documentation.
allowed-tools: Bash(date:*)
---

# Context
## Background
- Documentation goes into ~/docs
## Structure
### Background
- Documentation is an atomic document that includes a description of a problem or learning.
- The Documentation should assume no required domain knowledge.
### Document Structure
- Overview: Description of problem or lesson to be learned.
- Code Snippets: Minimal brief code snippets.
- Resolution: (If applicable) (optional) how the problem was resolved.
- Quiz Questions: Three questions to test understanding of the document.
- Rudiments: (If applicable) (optional) Coding challenges for a user to demonstrate hands-on understanding.
- Glossary: Glossary of important terms.
- Frontmatter: A property with a key of "Retention" and a numeric value from 1-10 with 10 meaning expert understanding of the content.
- Frontmatter: A property with a key of "Status" which starts with the status "READY".

# Instructions
- Review the provided context.
- Formulate the lesson.
- Choose a name which is a unix timestamp followed by a 3 word description.
- File name is is !`date +%s`-`<name>` that you generated in the last step.
- Save the filename to ~/.docs
