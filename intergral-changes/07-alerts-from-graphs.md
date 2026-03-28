# 7. Alerts from Graphs (Panel Menu)

**Problem:** Users want to create alert rules directly from a panel's data/queries without manually recreating them in the alerting UI.

**Solution:** Add a "New alert rule" option to the panel context menu (right-click menu). When clicked:
- Extract the panel's queries using the existing `scenesPanelToRuleFormValues()` utility
- Navigate to `/alerting/new` with the extracted values pre-populated in the form

**Implementation note:** The alert rule form schema may need to be relaxed (use a loose object schema for queries/expressions) to accommodate the variety of datasource query types that can come from panels.

## Where to implement

- **Panel menu behavior:** add the menu item with the extraction + navigation logic
- **Alert rule form schema:** ensure it accepts arbitrary query shapes from panels