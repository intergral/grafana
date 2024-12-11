import { css, cx } from '@emotion/css';
import { useEffect, useMemo } from 'react';
import { useLocation } from 'react-router-dom-v5-compat';

import { GrafanaTheme2, PageLayoutType } from '@grafana/data';
import { SceneComponentProps } from '@grafana/scenes';
import { useStyles2 } from '@grafana/ui';
import { TOP_BAR_LEVEL_HEIGHT } from 'app/core/components/AppChrome/types';
import NativeScrollbar from 'app/core/components/NativeScrollbar';
import { Page } from 'app/core/components/Page/Page';
import { EntityNotFound } from 'app/core/components/PageNotFound/EntityNotFound';
import { getNavModel } from 'app/core/selectors/navModel';
import DashboardEmpty from 'app/features/dashboard/dashgrid/DashboardEmpty';
import { useSelector } from 'app/types';

import { DashboardScene } from './DashboardScene';
import { NavToolbarActions } from './NavToolbarActions';
import { PanelSearchLayout } from './PanelSearchLayout';

export function DashboardSceneRenderer({ model }: SceneComponentProps<DashboardScene>) {
  const { controls, overlay, editview, editPanel, isEmpty, meta, viewPanelScene, panelSearch, panelsPerRow } =
    model.useState();
  const styles = useStyles2(getStyles);
  const location = useLocation();
  const navIndex = useSelector((state) => state.navIndex);
  const pageNav = model.getPageNav(location, navIndex);
  const bodyToRender = model.getBodyToRender();
  const navModel = getNavModel(navIndex, 'dashboards/browse');
  const hasControls = controls?.hasControls();
  const isSettingsOpen = editview !== undefined;

  useMemo(() => {
    if (viewPanelScene || isSettingsOpen || editPanel) {
      model.rememberScrollPos();
    }
  }, [isSettingsOpen, editPanel, viewPanelScene, model]);

  useEffect(() => {
    if (!viewPanelScene && !isSettingsOpen && !editPanel) {
      model.restoreScrollPos();
    }
  }, [isSettingsOpen, editPanel, viewPanelScene, model]);

  if (editview) {
    return (
      <>
        <editview.Component model={editview} />
        {overlay && <overlay.Component model={overlay} />}
      </>
    );
  }

  const notFound = meta.dashboardNotFound && <EntityNotFound entity="Dashboard" />;

  const emptyState = (
    <DashboardEmpty dashboard={model} canCreate={!!model.state.meta.canEdit} />
  );

  const withPanels = (
    <div className={cx(styles.body, !hasControls && styles.bodyWithoutControls)}>
      <bodyToRender.Component model={bodyToRender} />
    </div>
  );

  let body: React.ReactNode = [withPanels];

  if (notFound) {
    body = [notFound];
  } else if (isEmpty) {
    body = [emptyState, withPanels];
  } else if (panelSearch || panelsPerRow) {
    body = <PanelSearchLayout panelSearch={panelSearch} panelsPerRow={panelsPerRow} dashboard={model} />;
  }

  return (
    <Page navModel={navModel} pageNav={pageNav} layout={PageLayoutType.Custom}>
      {editPanel && <editPanel.Component model={editPanel} />}
      {!editPanel && (
        <NativeScrollbar divId="page-scrollbar" onSetScrollRef={model.onSetScrollRef}>
          <div className={cx(styles.pageContainer, hasControls && styles.pageContainerWithControls)}>
            <NavToolbarActions dashboard={model} />
            {controls && (
              <div className={styles.controlsWrapper}>
                <controls.Component model={controls} />
              </div>
            )}
            <div className={styles.canvasContent}>{body}</div>
          </div>
        </NativeScrollbar>
      )}
      {overlay && <overlay.Component model={overlay} />}
    </Page>
  );
}

function getStyles(theme: GrafanaTheme2) {
  return {
    pageContainer: css({
      display: 'grid',
      gridTemplateAreas: `"panels"`,
      gridTemplateColumns: '1fr',
      gridTemplateRows: '1fr',
      flexGrow: 1,
      marginTop: 0,
      backgroundColor: theme.colors.background.primary,
    }),
    pageContainerWithControls: css({
      gridTemplateAreas: `
        "controls"
        "panels"`,
      gridTemplateRows: 'auto 1fr',
    }),
    controlsWrapper: css({
      display: 'flex',
      flexDirection: 'column',
      gridArea: 'controls',
      padding: 0,
      margin: 0,
      position: 'sticky',
      top: TOP_BAR_LEVEL_HEIGHT,
      zIndex: theme.zIndex.navbarFixed - 1,
      borderBottom: `1px solid ${theme.colors.border.weak}`,
      backgroundColor: theme.colors.background.primary,
    }),
    canvasContent: css({
      display: 'flex',
      flexDirection: 'column',
      padding: theme.spacing(0.5, 2),
      flexBasis: '100%',
      gridArea: 'panels',
      flexGrow: 1,
      minWidth: 0,
    }),
    body: css({
      flexGrow: 1,
      display: 'flex',
      gap: '8px',
      paddingBottom: theme.spacing(2),
      boxSizing: 'border-box',
    }),
    bodyWithoutControls: css({
      paddingTop: theme.spacing(2),
    }),
  };
}
