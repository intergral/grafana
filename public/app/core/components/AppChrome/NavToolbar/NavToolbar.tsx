import { css } from '@emotion/css';
import * as React from 'react';

import { GrafanaTheme2, NavModelItem } from '@grafana/data';
import { Components } from '@grafana/e2e-selectors';
import { ToolbarButton, useStyles2 } from '@grafana/ui';
import { t } from 'app/core/internationalization';
import { HOME_NAV_ID } from 'app/core/reducers/navModel';
import { useSelector } from 'app/types';

import { Breadcrumbs } from '../../Breadcrumbs/Breadcrumbs';
import { buildBreadcrumbs } from '../../Breadcrumbs/utils';
import { TopSearchBarCommandPaletteTrigger } from '../TopBar/TopSearchBarCommandPaletteTrigger';
import { TOP_BAR_LEVEL_HEIGHT } from '../types';

export const TOGGLE_BUTTON_ID = 'mega-menu-toggle';

export interface Props {
  onToggleSearchBar(): void;
  onToggleMegaMenu(): void;
  onToggleKioskMode(): void;
  searchBarHidden?: boolean;
  sectionNav: NavModelItem;
  pageNav?: NavModelItem;
  actions: React.ReactNode;
}

export function NavToolbar({
  actions,
  searchBarHidden,
  sectionNav,
  pageNav,
  onToggleKioskMode,
}: Props) {
  const homeNav = useSelector((state) => state.navIndex)[HOME_NAV_ID];
  const styles = useStyles2(getStyles);
  const breadcrumbs = buildBreadcrumbs(sectionNav, pageNav, homeNav);

  return (
    <div data-testid={Components.NavToolbar.container} className={styles.pageToolbar}>
      <Breadcrumbs breadcrumbs={breadcrumbs} className={styles.breadcrumbsWrapper} />
      <div className={styles.actions}>
        <div className={styles.commandPalette}>
          <TopSearchBarCommandPaletteTrigger/>
        </div>
        {actions}
        {searchBarHidden && (
          <ToolbarButton
            onClick={onToggleKioskMode}
            narrow
            title={t('navigation.toolbar.enable-kiosk', 'Enable kiosk mode')}
            icon="monitor"
          />
        )}
      </div>
    </div>
  );
}

const getStyles = (theme: GrafanaTheme2) => {
  return {
    breadcrumbsWrapper: css({
      display: 'flex',
      overflow: 'hidden',
      [theme.breakpoints.down('sm')]: {
        minWidth: '50%',
      },
    }),
    pageToolbar: css({
      height: TOP_BAR_LEVEL_HEIGHT,
      display: 'flex',
      padding: theme.spacing(0, 1, 0, 2),
      alignItems: 'center',
      borderBottom: `1px solid ${theme.colors.border.weak}`,
    }),
    menuButton: css({
      display: 'flex',
      alignItems: 'center',
      marginRight: theme.spacing(1),
    }),
    commandPalette: css({
      width: '100%',
      maxWidth: '550px',
      display: 'flex',
      alignItems: 'center',

      [theme.breakpoints.down('sm')]: {
        maxWidth: 'none',
      },
    }),
    actions: css({
      label: 'NavToolbar-actions',
      display: 'flex',
      alignItems: 'center',
      flexWrap: 'nowrap',
      justifyContent: 'flex-end',
      paddingLeft: theme.spacing(1),
      flexGrow: 1,
      gap: theme.spacing(1),
      minWidth: 0,

      '.body-drawer-open &': {
        display: 'none',
      },
    }),
  };
};
