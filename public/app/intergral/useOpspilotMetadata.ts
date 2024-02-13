import domtoimage from 'dom-to-image';
import { useEffect } from 'react';

import { getTimeSrv } from 'app/features/dashboard/services/TimeSrv';


export const useOpspilotMetadata = () => {
    useEffect(() => {
        const event = async (event: MessageEvent) => {
          if (event.data.type === 'opspilot-host.getMetadata') {
            const screenshot = await domtoimage.toPng(document.querySelector('main') as HTMLElement, { bgcolor: '#000' });
            window.parent.postMessage({type: "opspilot-slave.sendMetadata", metadata: { 
              url: window.location.pathname,
              timeStart: getTimeSrv().timeRange().from.valueOf(), 
              timeEnd: getTimeSrv().timeRange().to.valueOf(),
              timezone: getTimeSrv().timeModel?.getTimezone() === 'browser' ? Intl.DateTimeFormat().resolvedOptions().timeZone : getTimeSrv().timeModel?.getTimezone(),
              screenshot: screenshot
            } }, '*');
          }
        };
        window.addEventListener('message', event);
        return () => {
          window.removeEventListener('message', event);
        }
      }, []);

}
