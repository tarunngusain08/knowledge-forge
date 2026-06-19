# UI Screenshot Inventory

No product UI screenshots are committed in this refresh.

The documentation request explicitly required real application views and no
fabricated screenshots. This checkout does not currently contain runnable web
dependencies (`ui/web/node_modules` is absent), the Streamlit dependency is not
installed in the default local Python environment, and no seeded live backend
session was created during this docs-only branch.

Future screenshots should be captured only from a real local or seeded-demo
Knowledge Forge session.

Recommended capture list:

- home or login page
- repository registration and indexing flow
- repository Q&A flow
- grounded answer with citations
- retrieval trace or debug evidence view
- Deep-Dive Report output
- evaluation or benchmark report UI if available

Do not place mocked product screens in this directory.
