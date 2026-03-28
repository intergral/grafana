# 3. Ad-hoc Filter Variable UID Resolution

> **Note:** Check if upstream has fixed this in the target version before reimplementing.

**Problem:** When converting dashboard variables from the v2 schema to scenes, ad-hoc filter variables sometimes have the datasource *name* stored in the UID field. The DrilldownDependenciesManager then fails to match variables because it compares by resolved UID.

**Solution:** When creating `AdHocFiltersVariable` instances from saved models, resolve the datasource UID by looking it up via `getDataSourceSrv().getInstanceSettings()` before constructing the variable. This handles both the normal creation path and the snapshot creation path.

## Where to implement

- **Dashboard scene serialization** (v2-to-scene transform): resolve datasource UID before creating AdHocFiltersVariable
- **Variable creation utilities:** same resolution in `createVariablesForSnapshot` and `createSceneVariableFromVariableModel`