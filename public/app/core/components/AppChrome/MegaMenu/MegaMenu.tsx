import { DOMAttributes } from '@react-types/shared';
import { memo, forwardRef} from 'react';

import { NavModelItem } from '@grafana/data';
import { config } from '@grafana/runtime';
import { useGrafana } from 'app/core/context/GrafanaContext';
import {  useSelector } from 'app/types';

import { usePinnedItems } from './hooks';
import { enrichWithInteractionTracking, findByUrl } from './utils';

export const MENU_WIDTH = '300px';

export interface Props extends DOMAttributes {
  onClose: () => void;
}

export const MegaMenu = memo(
  forwardRef<HTMLDivElement, Props>(({ onClose, ...restProps }, ref) => {
    const navTree = useSelector((state) => state.navBarTree);
    const { chrome } = useGrafana();
    const state = chrome.useState();
    const pinnedItems = usePinnedItems();

    // Remove profile + help from tree
    const navItems = navTree
      .filter((item) => item.id !== 'profile' && item.id !== 'help')
      .map((item) => enrichWithInteractionTracking(item, state.megaMenuDocked));

    if (config.featureToggles.pinNavItems) {
      const bookmarksItem = findByUrl(navItems, '/bookmarks');
      if (bookmarksItem) {
        // Add children to the bookmarks section
        bookmarksItem.children = pinnedItems.reduce((acc: NavModelItem[], url) => {
          const item = findByUrl(navItems, url);
          if (!item) {
            return acc;
          }
          const newItem = {
            id: item.id,
            text: item.text,
            url: item.url,
            parentItem: { id: 'bookmarks', text: 'Bookmarks' },
          };
          acc.push(enrichWithInteractionTracking(newItem, state.megaMenuDocked));
          return acc;
        }, []);
      }
    }

    return null
  })
);

MegaMenu.displayName = 'MegaMenu';
