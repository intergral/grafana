import { useEffect } from 'react';

import { locationService } from '@grafana/runtime';

/**
 * Hook that listens for postMessage events from the parent window to enable
 * navigation without full iframe reloads when embedded.
 */
export const useIframeNavigation = () => {
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.data?.type === 'navigate' && typeof event.data?.path === 'string') {
        locationService.push(event.data.path);
      }
    };

    window.addEventListener('message', handleMessage);
    return () => {
      window.removeEventListener('message', handleMessage);
    };
  }, []);
};
