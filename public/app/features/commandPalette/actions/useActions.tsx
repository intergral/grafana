import { useRegisterActions } from 'kbar';
import { useEffect, useState } from 'react';

import { CommandPaletteAction } from '../types';

import { getRecentDashboardActions } from './dashboardActions';
import { useStaticActions } from './staticActions';

/**
 * Register navigation actions to different parts of grafana or some preferences stuff like themes.
 */
export function useRegisterStaticActions() {
  const staticActions = useStaticActions();

  useRegisterActions(staticActions, [staticActions]);
}

export function useRegisterRecentDashboardsActions() {
  const [recentDashboardActions, setRecentDashboardActions] = useState<CommandPaletteAction[]>([]);
  useEffect(() => {
    getRecentDashboardActions()
      .then((recentDashboardActions) => setRecentDashboardActions(recentDashboardActions))
      .catch((err) => {
        console.error('Error loading recent dashboard actions', err);
      });
  }, []);

  useRegisterActions(recentDashboardActions, [recentDashboardActions]);
}
