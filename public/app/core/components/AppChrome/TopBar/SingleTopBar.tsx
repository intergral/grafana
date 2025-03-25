import { css } from '@emotion/css';
import { memo } from 'react';

import { GrafanaTheme2, NavModelItem } from '@grafana/data';
import { Stack, useStyles2 } from '@grafana/ui';
import { useGrafana } from 'app/core/context/GrafanaContext';
import { HOME_NAV_ID } from 'app/core/reducers/navModel';
import { useSelector } from 'app/types';

import { Breadcrumbs } from '../../Breadcrumbs/Breadcrumbs';
import { buildBreadcrumbs } from '../../Breadcrumbs/utils';
import { TOP_BAR_LEVEL_HEIGHT } from '../types';


import { TopSearchBarCommandPaletteTrigger } from './TopSearchBarCommandPaletteTrigger';

export const MEGA_MENU_TOGGLE_ID = 'mega-menu-toggle';

interface Props {
  sectionNav: NavModelItem;
  pageNav?: NavModelItem;
  onToggleMegaMenu(): void;
  onToggleKioskMode(): void;
}

export const SingleTopBar = memo(function SingleTopBar({
  pageNav,
  sectionNav,
}: Props) {
  const { chrome } = useGrafana();
  const state = chrome.useState();
  const menuDockedAndOpen = !state.chromeless && state.megaMenuDocked && state.megaMenuOpen;
  const styles = useStyles2(getStyles, menuDockedAndOpen);

  const homeNav = useSelector((state) => state.navIndex)[HOME_NAV_ID];
  const breadcrumbs = buildBreadcrumbs(sectionNav, pageNav, homeNav);

  return (
    <div className={styles.layout}>
      <Stack minWidth={0} gap={0.5} alignItems="center">
        <Breadcrumbs breadcrumbs={breadcrumbs} className={styles.breadcrumbsWrapper} />
      </Stack>
      <Stack gap={0.5} alignItems="center">
        <TopSearchBarCommandPaletteTrigger />
      </Stack>
    </div>
  );
});

const getStyles = (theme: GrafanaTheme2, menuDockedAndOpen: boolean) => ({
  layout: css({
    height: TOP_BAR_LEVEL_HEIGHT,
    display: 'flex',
    gap: theme.spacing(2),
    alignItems: 'center',
    padding: theme.spacing(0, 1),
    paddingLeft: menuDockedAndOpen ? theme.spacing(3.5) : theme.spacing(0.75),
    borderBottom: `1px solid ${theme.colors.border.weak}`,
    justifyContent: 'space-between',

    [theme.breakpoints.up('lg')]: {
      gridTemplateColumns: '2fr minmax(550px, 1fr)',
      display: 'grid',
      justifyContent: 'flex-start',
    },
  }),
  breadcrumbsWrapper: css({
    display: 'flex',
    overflow: 'hidden',
    [theme.breakpoints.down('sm')]: {
      minWidth: '40%',
    },
  }),
  img: css({
    alignSelf: 'center',
    height: theme.spacing(3),
    width: theme.spacing(3),
  }),
  profileButton: css({
    padding: theme.spacing(0, 0.5),
    img: {
      borderRadius: theme.shape.radius.circle,
      height: '24px',
      marginRight: 0,
      width: '24px',
    },
  }),
  kioskToggle: css({
    [theme.breakpoints.down('lg')]: {
      display: 'none',
    },
  }),
});
